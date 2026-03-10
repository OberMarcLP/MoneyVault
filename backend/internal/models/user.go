package models

import (
	"encoding/json"
	"fmt"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleViewer UserRole = "viewer"
)

type User struct {
	ID                  uuid.UUID       `db:"id" json:"id"`
	Email               string          `db:"email" json:"email"`
	PasswordHash        string          `db:"password_hash" json:"-"`
	Role                UserRole        `db:"role" json:"role"`
	TOTPSecret          *string         `db:"totp_secret" json:"-"`
	TOTPEnabled         bool            `db:"totp_enabled" json:"totp_enabled"`
	EncryptedDEK        string          `db:"encrypted_dek" json:"-"`
	KEKSalt             string          `db:"kek_salt" json:"-"`
	Preferences         json.RawMessage `db:"preferences" json:"preferences" swaggertype:"object"`
	EmailVerified       bool            `db:"email_verified" json:"email_verified"`
	FailedLoginAttempts int             `db:"failed_login_attempts" json:"-"`
	LockedUntil         *time.Time      `db:"locked_until" json:"-"`
	E2EEnabled          bool            `db:"e2e_enabled" json:"e2e_enabled"`
	E2EEncryptedDEK     string          `db:"e2e_encrypted_dek" json:"-"`
	E2EKEKSalt          string          `db:"e2e_kek_salt" json:"-"`
	CreatedAt           time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
}

type UserPreferences struct {
	Theme               string `json:"theme"`
	Currency            string `json:"currency"`
	Locale              string `json:"locale"`
	OnboardingDismissed bool   `json:"onboarding_dismissed,omitempty"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TOTPCode string `json:"totp_code,omitempty"`
}

type LoginResponse struct {
	AccessToken     string `json:"access_token"`
	User            User   `json:"user"`
	E2EEncryptedDEK string `json:"e2e_encrypted_dek,omitempty"`
	E2EKEKSalt      string `json:"e2e_kek_salt,omitempty"`
}

type EnableE2ERequest struct {
	Password        string `json:"password" binding:"required"`
	E2EEncryptedDEK string `json:"e2e_encrypted_dek" binding:"required"`
	E2EKEKSalt      string `json:"e2e_kek_salt" binding:"required"`
}

type E2EMigrateDataRequest struct {
	Accounts     []E2EMigrateAccount     `json:"accounts"`
	Transactions []E2EMigrateTransaction `json:"transactions"`
}

type E2EMigrateAccount struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Balance string `json:"balance"`
}

type E2EMigrateTransaction struct {
	ID          string `json:"id"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
}

type TOTPSetupResponse struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
	QR     string `json:"qr"`
}

type TOTPVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type UpdatePreferencesRequest struct {
	Theme               *string `json:"theme"`
	Currency            *string `json:"currency"`
	Locale              *string `json:"locale"`
	OnboardingDismissed *bool   `json:"onboarding_dismissed,omitempty"`
}

func ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
