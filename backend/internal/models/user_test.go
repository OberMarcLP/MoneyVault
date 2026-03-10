package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePasswordComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{"valid password", "StrongP@ss1", ""},
		{"valid complex", "MyP@ssw0rd!", ""},
		{"too short", "Aa1!", "at least 8"},
		{"no uppercase", "password1!", "uppercase"},
		{"no lowercase", "PASSWORD1!", "lowercase"},
		{"no digit", "Password!", "digit"},
		{"no special", "Password1", "special"},
		{"empty", "", "at least 8"},
		{"just length", "abcdefgh", "uppercase"},
		{"unicode special", "Password1€", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordComplexity(tt.password)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordComplexity_MinLength(t *testing.T) {
	err := ValidatePasswordComplexity("Aa1!xyz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8")

	err = ValidatePasswordComplexity("Aa1!xyzz")
	assert.NoError(t, err)
}
