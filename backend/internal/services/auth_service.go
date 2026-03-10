package services

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"moneyvault/internal/config"
	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

type AuthService struct {
	userRepo       *repositories.UserRepository
	catRepo        *repositories.CategoryRepository
	tokenRepo      *repositories.TokenRepository
	dekSessionRepo *repositories.DEKSessionRepository
	enc            *encryption.Service
	cfg            *config.Config
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	catRepo *repositories.CategoryRepository,
	tokenRepo *repositories.TokenRepository,
	dekSessionRepo *repositories.DEKSessionRepository,
	enc *encryption.Service,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		catRepo:        catRepo,
		tokenRepo:      tokenRepo,
		dekSessionRepo: dekSessionRepo,
		enc:            enc,
		cfg:            cfg,
	}
}

type TokenClaims struct {
	UserID uuid.UUID       `json:"user_id"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(req models.CreateUserRequest) (*models.User, string, error) {
	if err := models.ValidatePasswordComplexity(req.Password); err != nil {
		return nil, "", err
	}

	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, "", errors.New("email already registered")
	}

	salt, err := encryption.GenerateSalt()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate salt: %w", err)
	}

	passwordHash := hashPassword(req.Password, salt)

	dek, err := s.enc.GenerateDEK()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate DEK: %w", err)
	}

	kek := s.enc.DeriveKEK(req.Password, salt)
	encryptedDEK, err := s.enc.EncryptDEK(dek, kek)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	role := models.RoleUser
	count, _ := s.userRepo.Count()
	if count == 0 {
		role = models.RoleAdmin
	}

	defaultPrefs, _ := json.Marshal(models.UserPreferences{
		Theme:    "system",
		Currency: "USD",
		Locale:   "en-US",
	})

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: base64.StdEncoding.EncodeToString(passwordHash),
		Role:         role,
		TOTPEnabled:  false,
		EncryptedDEK: encryptedDEK,
		KEKSalt:      base64.StdEncoding.EncodeToString(salt),
		Preferences:  defaultPrefs,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	s.enc.StoreDEK(user.ID, dek)
	s.persistDEKSession(user.ID, dek)

	if err := s.catRepo.CreateDefaults(user.ID); err != nil {
		return nil, "", fmt.Errorf("failed to create default categories: %w", err)
	}

	return user, "", nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, string, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Check account lockout
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remaining := time.Until(*user.LockedUntil).Minutes()
		return nil, "", fmt.Errorf("account temporarily locked, try again in %.0f minutes", remaining+1)
	}

	salt, err := base64.StdEncoding.DecodeString(user.KEKSalt)
	if err != nil {
		return nil, "", errors.New("internal error")
	}

	passwordHash := hashPassword(req.Password, salt)
	storedHash, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		return nil, "", errors.New("internal error")
	}

	if !compareHashes(passwordHash, storedHash) {
		// Increment failed attempts and possibly lock
		attempts, _ := s.userRepo.IncrementFailedAttempts(user.ID)
		if attempts >= 10 {
			_ = s.userRepo.LockUser(user.ID, time.Now().Add(1*time.Hour))
			return nil, "", errors.New("account locked for 1 hour due to too many failed attempts")
		} else if attempts >= 5 {
			_ = s.userRepo.LockUser(user.ID, time.Now().Add(15*time.Minute))
			return nil, "", errors.New("account locked for 15 minutes due to too many failed attempts")
		}
		return nil, "", errors.New("invalid credentials")
	}

	if user.TOTPEnabled {
		if req.TOTPCode == "" {
			return nil, "", errors.New("totp_required")
		}
		if user.TOTPSecret == nil {
			return nil, "", errors.New("internal error")
		}
		if !totp.Validate(req.TOTPCode, *user.TOTPSecret) {
			return nil, "", errors.New("invalid TOTP code")
		}
	}

	// Reset failed attempts on successful login
	_ = s.userRepo.ResetFailedAttempts(user.ID)

	kek := s.enc.DeriveKEK(req.Password, salt)
	dek, err := s.enc.DecryptDEK(user.EncryptedDEK, kek)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decrypt DEK: %w", err)
	}
	s.enc.StoreDEK(user.ID, dek)
	s.persistDEKSession(user.ID, dek)

	// Track E2E status in encryption service
	if user.E2EEnabled {
		s.enc.SetE2EUser(user.ID, true)
	}

	accessToken, err := s.generateToken(user, s.cfg.AccessTokenExpiry)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.generateToken(user, s.cfg.RefreshTokenExpiry)
	if err != nil {
		return nil, "", err
	}

	resp := &models.LoginResponse{
		AccessToken: accessToken,
		User:        *user,
	}

	// Include E2E DEK data so client can derive KEK and decrypt DEK
	if user.E2EEnabled {
		resp.E2EEncryptedDEK = user.E2EEncryptedDEK
		resp.E2EKEKSalt = user.E2EKEKSalt
	}

	return resp, refreshToken, nil
}

func (s *AuthService) RefreshToken(refreshTokenStr string) (string, error) {
	claims, err := s.ValidateToken(refreshTokenStr)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Check if token has been revoked
	if s.tokenRepo.IsRevoked(hashToken(refreshTokenStr)) {
		return "", errors.New("token has been revoked")
	}

	if _, err := s.enc.GetDEK(claims.UserID); err != nil {
		// Try to restore DEK from persisted session
		if !s.restoreDEKSession(claims.UserID) {
			return "", errors.New("session expired, please login again")
		}
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", errors.New("user not found")
	}

	return s.generateToken(user, s.cfg.AccessTokenExpiry)
}

func (s *AuthService) SetupTOTP(userID uuid.UUID) (*models.TOTPSetupResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "MoneyVault",
		AccountName: user.Email,
	})
	if err != nil {
		return nil, err
	}

	user.TOTPSecret = strPtr(key.Secret())
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &models.TOTPSetupResponse{
		Secret: key.Secret(),
		URL:    key.URL(),
	}, nil
}

func (s *AuthService) VerifyTOTP(userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user.TOTPSecret == nil {
		return errors.New("TOTP not set up")
	}
	if !totp.Validate(code, *user.TOTPSecret) {
		return errors.New("invalid TOTP code")
	}
	user.TOTPEnabled = true
	return s.userRepo.Update(user)
}

func (s *AuthService) DisableTOTP(userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	user.TOTPEnabled = false
	user.TOTPSecret = nil
	return s.userRepo.Update(user)
}

func (s *AuthService) ValidateToken(tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) GetUser(id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *AuthService) UpdatePreferences(userID uuid.UUID, req models.UpdatePreferencesRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	var prefs models.UserPreferences
	if err := json.Unmarshal(user.Preferences, &prefs); err != nil {
		prefs = models.UserPreferences{Theme: "system", Currency: "USD", Locale: "en-US"}
	}
	if req.Theme != nil {
		prefs.Theme = *req.Theme
	}
	if req.Currency != nil {
		prefs.Currency = *req.Currency
	}
	if req.Locale != nil {
		prefs.Locale = *req.Locale
	}
	if req.OnboardingDismissed != nil {
		prefs.OnboardingDismissed = *req.OnboardingDismissed
	}

	user.Preferences, _ = json.Marshal(prefs)
	return s.userRepo.Update(user)
}

func (s *AuthService) Logout(userID uuid.UUID, refreshToken string) {
	s.enc.RemoveDEK(userID)
	if s.dekSessionRepo != nil {
		_ = s.dekSessionRepo.Delete(userID)
	}
	if refreshToken != "" {
		claims, err := s.ValidateToken(refreshToken)
		if err == nil {
			_ = s.tokenRepo.RevokeToken(hashToken(refreshToken), claims.ExpiresAt.Time)
		}
	}
}

func (s *AuthService) VerifyEmail(userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	user.EmailVerified = true
	return s.userRepo.Update(user)
}

func (s *AuthService) CleanupExpiredTokens() (int64, error) {
	return s.tokenRepo.CleanupExpired()
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) generateToken(user *models.User, expiry time.Duration) (string, error) {
	claims := TokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func hashPassword(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
}

func compareHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

func strPtr(s string) *string {
	return &s
}

// GenerateAccessToken creates a JWT access token for the given user.
func (s *AuthService) GenerateAccessToken(userID uuid.UUID, role models.UserRole) (string, error) {
	user := &models.User{ID: userID, Role: role}
	return s.generateToken(user, s.cfg.AccessTokenExpiry)
}

// GenerateRefreshToken creates a JWT refresh token for the given user.
func (s *AuthService) GenerateRefreshToken(userID uuid.UUID, role models.UserRole) (string, error) {
	user := &models.User{ID: userID, Role: role}
	return s.generateToken(user, s.cfg.RefreshTokenExpiry)
}

// RestoreDEKSession loads a persisted DEK from the database into memory for passkey login.
func (s *AuthService) RestoreDEKSession(userID uuid.UUID) bool {
	return s.restoreDEKSession(userID)
}

// RefreshTokenExpiry returns the configured refresh token expiry duration.
func (s *AuthService) RefreshTokenExpiry() time.Duration {
	return s.cfg.RefreshTokenExpiry
}

// persistDEKSession encrypts and saves the DEK to the database for crash resilience.
func (s *AuthService) persistDEKSession(userID uuid.UUID, dek []byte) {
	if s.dekSessionRepo == nil {
		return
	}
	encDEK, err := s.enc.EncryptDEKForSession(dek)
	if err != nil {
		return
	}
	_ = s.dekSessionRepo.Upsert(&repositories.DEKSession{
		UserID:       userID,
		EncryptedDEK: encDEK,
		ExpiresAt:    time.Now().Add(s.cfg.RefreshTokenExpiry),
	})
}

// restoreDEKSession loads a persisted DEK session from the database into memory.
func (s *AuthService) restoreDEKSession(userID uuid.UUID) bool {
	if s.dekSessionRepo == nil {
		return false
	}
	session, err := s.dekSessionRepo.GetByUserID(userID)
	if err != nil {
		return false
	}
	dek, err := s.enc.DecryptDEKFromSession(session.EncryptedDEK)
	if err != nil {
		return false
	}
	s.enc.StoreDEK(userID, dek)
	return true
}

// RestoreAllSessions loads all active DEK sessions from the database on startup.
func (s *AuthService) RestoreAllSessions() int {
	if s.dekSessionRepo == nil {
		return 0
	}
	sessions, err := s.dekSessionRepo.GetAllActive()
	if err != nil {
		return 0
	}
	count := 0
	for _, session := range sessions {
		dek, err := s.enc.DecryptDEKFromSession(session.EncryptedDEK)
		if err != nil {
			continue
		}
		s.enc.StoreDEK(session.UserID, dek)
		count++
	}
	return count
}

// CleanupExpiredDEKSessions removes expired DEK sessions from the database.
func (s *AuthService) CleanupExpiredDEKSessions() (int64, error) {
	if s.dekSessionRepo == nil {
		return 0, nil
	}
	return s.dekSessionRepo.CleanupExpired()
}
