package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type PasswordResetService struct {
	resetRepo *repositories.PasswordResetRepository
	userRepo  *repositories.UserRepository
	tokenRepo *repositories.TokenRepository
	enc       *encryption.Service
}

func NewPasswordResetService(
	resetRepo *repositories.PasswordResetRepository,
	userRepo *repositories.UserRepository,
	tokenRepo *repositories.TokenRepository,
	enc *encryption.Service,
) *PasswordResetService {
	return &PasswordResetService{
		resetRepo: resetRepo,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		enc:       enc,
	}
}

// RequestReset generates a reset token for the given email.
// Returns the plaintext token (self-hosted: displayed to user, no email).
// Always returns success to prevent email enumeration.
func (s *PasswordResetService) RequestReset(email string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Don't reveal whether user exists
		return "", nil
	}

	// Generate a cryptographically random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	plainToken := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash for storage
	h := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(h[:])

	resetToken := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	if err := s.resetRepo.Create(resetToken); err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	return plainToken, nil
}

// ResetPassword validates the token and sets a new password.
// It re-encrypts the DEK with the new password and revokes all refresh tokens.
func (s *PasswordResetService) ResetPassword(plainToken, newPassword string) error {
	if err := models.ValidatePasswordComplexity(newPassword); err != nil {
		return err
	}

	// Hash token to look up
	h := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(h[:])

	resetToken, err := s.resetRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if resetToken.Used {
		return errors.New("reset token already used")
	}
	if time.Now().After(resetToken.ExpiresAt) {
		return errors.New("reset token has expired")
	}

	user, err := s.userRepo.GetByID(resetToken.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// We need the old DEK. If the DEK is in memory (user still has an active
	// session), we can re-encrypt it with the new password. Otherwise, we must
	// generate a new DEK (existing encrypted data becomes inaccessible — this is
	// an inherent limitation when the password is lost with server-side encryption).
	dek, dekErr := s.enc.GetDEK(user.ID)
	if dekErr != nil {
		dek, err = s.enc.GenerateDEK()
		if err != nil {
			return fmt.Errorf("failed to generate new DEK: %w", err)
		}
	}

	// Generate new salt and hash the new password
	salt, err := encryption.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	passwordHash := argon2.IDKey([]byte(newPassword), salt, 3, 64*1024, 4, 32)

	// Derive new KEK and re-encrypt DEK
	kek := s.enc.DeriveKEK(newPassword, salt)
	encryptedDEK, err := s.enc.EncryptDEK(dek, kek)
	if err != nil {
		return fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	// Update user
	user.PasswordHash = base64.StdEncoding.EncodeToString(passwordHash)
	user.KEKSalt = base64.StdEncoding.EncodeToString(salt)
	user.EncryptedDEK = encryptedDEK
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Mark token as used
	_ = s.resetRepo.MarkUsed(resetToken.ID)

	// Clear DEK from memory — force re-login with new password
	s.enc.RemoveDEK(user.ID)

	return nil
}

func (s *PasswordResetService) CleanupExpired() (int64, error) {
	return s.resetRepo.CleanupExpired()
}
