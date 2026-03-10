package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"moneyvault/internal/config"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"
)

type WebAuthnService struct {
	wa           *webauthn.WebAuthn
	repo         *repositories.WebAuthnRepository
	userRepo     *repositories.UserRepository
	authService  *AuthService
	challenges   map[string]*webauthn.SessionData
	challengesMu sync.RWMutex
}

func NewWebAuthnService(
	cfg *config.Config,
	repo *repositories.WebAuthnRepository,
	userRepo *repositories.UserRepository,
	authService *AuthService,
) (*WebAuthnService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: "MoneyVault",
		RPID:          cfg.WebAuthnRPID,
		RPOrigins:     cfg.WebAuthnRPOrigins,
	}

	wa, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn: %w", err)
	}

	s := &WebAuthnService{
		wa:          wa,
		repo:        repo,
		userRepo:    userRepo,
		authService: authService,
		challenges:  make(map[string]*webauthn.SessionData),
	}

	go s.cleanupChallenges()

	return s, nil
}

// webauthnUser wraps our User model to implement the webauthn.User interface.
type webauthnUser struct {
	user  *models.User
	creds []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte {
	id := u.user.ID
	return id[:]
}

func (u *webauthnUser) WebAuthnName() string {
	return u.user.Email
}

func (u *webauthnUser) WebAuthnDisplayName() string {
	return u.user.Email
}

func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.creds
}

func (s *WebAuthnService) toWebAuthnUser(user *models.User) (*webauthnUser, error) {
	dbCreds, err := s.repo.ListByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	creds := make([]webauthn.Credential, len(dbCreds))
	for i, c := range dbCreds {
		transports := make([]protocol.AuthenticatorTransport, len(c.Transport))
		for j, t := range c.Transport {
			transports[j] = protocol.AuthenticatorTransport(t)
		}
		creds[i] = webauthn.Credential{
			ID:              c.CredentialID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Transport:       transports,
			Authenticator: webauthn.Authenticator{
				AAGUID:    c.AAGUID,
				SignCount: uint32(c.SignCount),
			},
		}
	}

	return &webauthnUser{user: user, creds: creds}, nil
}

func (s *WebAuthnService) storeChallenge(key string, session *webauthn.SessionData) {
	s.challengesMu.Lock()
	defer s.challengesMu.Unlock()
	s.challenges[key] = session
}

func (s *WebAuthnService) getChallenge(key string) (*webauthn.SessionData, bool) {
	s.challengesMu.Lock()
	defer s.challengesMu.Unlock()
	session, ok := s.challenges[key]
	if ok {
		delete(s.challenges, key)
	}
	return session, ok
}

func (s *WebAuthnService) cleanupChallenges() {
	// Challenges auto-expire when retrieved. This just cleans stale ones.
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.challengesMu.Lock()
		// Clear all — challenges older than 5 minutes are stale anyway
		s.challenges = make(map[string]*webauthn.SessionData)
		s.challengesMu.Unlock()
	}
}

// BeginRegistration starts the WebAuthn credential registration for an authenticated user.
func (s *WebAuthnService) BeginRegistration(userID uuid.UUID) (*protocol.CredentialCreation, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	waUser, err := s.toWebAuthnUser(user)
	if err != nil {
		return nil, err
	}

	// Exclude existing credentials to prevent re-registration
	excludeList := make([]protocol.CredentialDescriptor, len(waUser.creds))
	for i, c := range waUser.creds {
		excludeList[i] = protocol.CredentialDescriptor{
			Type:            protocol.PublicKeyCredentialType,
			CredentialID:    c.ID,
		}
	}

	options, session, err := s.wa.BeginRegistration(
		waUser,
		webauthn.WithExclusions(excludeList),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
	)
	if err != nil {
		return nil, fmt.Errorf("begin registration failed: %w", err)
	}

	s.storeChallenge("reg:"+userID.String(), session)
	return options, nil
}

// FinishRegistration completes the WebAuthn credential registration.
func (s *WebAuthnService) FinishRegistration(userID uuid.UUID, name string, response *protocol.ParsedCredentialCreationData) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	waUser, err := s.toWebAuthnUser(user)
	if err != nil {
		return err
	}

	session, ok := s.getChallenge("reg:" + userID.String())
	if !ok {
		return fmt.Errorf("registration challenge expired, please try again")
	}

	credential, err := s.wa.CreateCredential(waUser, *session, response)
	if err != nil {
		return fmt.Errorf("credential verification failed: %w", err)
	}

	if name == "" {
		name = "My Passkey"
	}

	transports := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transports[i] = string(t)
	}

	dbCred := &models.WebAuthnCredential{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            name,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       transports,
		SignCount:       int(credential.Authenticator.SignCount),
		AAGUID:          credential.Authenticator.AAGUID,
	}

	return s.repo.Create(dbCred)
}

// BeginLogin starts the WebAuthn authentication flow.
func (s *WebAuthnService) BeginLogin(email string) (*protocol.CredentialAssertion, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remaining := time.Until(*user.LockedUntil).Minutes()
		return nil, fmt.Errorf("account temporarily locked, try again in %.0f minutes", remaining)
	}

	waUser, err := s.toWebAuthnUser(user)
	if err != nil {
		return nil, err
	}

	if len(waUser.creds) == 0 {
		return nil, fmt.Errorf("no passkeys registered for this account")
	}

	options, session, err := s.wa.BeginLogin(waUser)
	if err != nil {
		return nil, fmt.Errorf("begin login failed: %w", err)
	}

	s.storeChallenge("login:"+user.ID.String(), session)
	return options, nil
}

// FinishLogin completes the WebAuthn authentication and returns tokens.
func (s *WebAuthnService) FinishLogin(email string, response *protocol.ParsedCredentialAssertionData) (string, string, *models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", "", nil, fmt.Errorf("invalid credentials")
	}

	waUser, err := s.toWebAuthnUser(user)
	if err != nil {
		return "", "", nil, err
	}

	session, ok := s.getChallenge("login:" + user.ID.String())
	if !ok {
		return "", "", nil, fmt.Errorf("login challenge expired, please try again")
	}

	credential, err := s.wa.ValidateLogin(waUser, *session, response)
	if err != nil {
		s.userRepo.IncrementFailedAttempts(user.ID)
		if user.FailedLoginAttempts+1 >= 10 {
			s.userRepo.LockUser(user.ID, time.Now().Add(1*time.Hour))
		} else if user.FailedLoginAttempts+1 >= 5 {
			s.userRepo.LockUser(user.ID, time.Now().Add(5*time.Minute))
		}
		return "", "", nil, fmt.Errorf("authentication failed")
	}

	// Update sign count
	dbCred, err := s.repo.GetByCredentialID(credential.ID)
	if err == nil {
		s.repo.UpdateSignCount(dbCred.ID, int(credential.Authenticator.SignCount))
	}

	// Reset failed attempts
	s.userRepo.ResetFailedAttempts(user.ID)

	// Restore DEK from session persistence
	if !s.authService.RestoreDEKSession(user.ID) {
		return "", "", nil, fmt.Errorf("session expired, please login with password to re-establish encryption keys")
	}

	// Generate tokens
	accessToken, err := s.authService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := s.authService.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

// ListCredentials returns all WebAuthn credentials for a user.
func (s *WebAuthnService) ListCredentials(userID uuid.UUID) ([]models.WebAuthnCredential, error) {
	return s.repo.ListByUserID(userID)
}

// DeleteCredential removes a WebAuthn credential.
func (s *WebAuthnService) DeleteCredential(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
