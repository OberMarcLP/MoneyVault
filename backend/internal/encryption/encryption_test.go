package encryption

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.keys)
	assert.NotNil(t, svc.e2eUsers)
}

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt()
	require.NoError(t, err)
	assert.Len(t, salt, 16)

	// Two salts should be different (random)
	salt2, err := GenerateSalt()
	require.NoError(t, err)
	assert.NotEqual(t, salt, salt2)
}

func TestGenerateDEK(t *testing.T) {
	svc := NewService()
	dek, err := svc.GenerateDEK()
	require.NoError(t, err)
	assert.Len(t, dek, 32)

	dek2, err := svc.GenerateDEK()
	require.NoError(t, err)
	assert.NotEqual(t, dek, dek2)
}

func TestDeriveKEK(t *testing.T) {
	svc := NewService()
	salt := []byte("testsalt12345678")

	kek := svc.DeriveKEK("password123", salt)
	assert.Len(t, kek, 32)

	// Same password + salt => same KEK
	kek2 := svc.DeriveKEK("password123", salt)
	assert.Equal(t, kek, kek2)

	// Different password => different KEK
	kek3 := svc.DeriveKEK("differentpass", salt)
	assert.NotEqual(t, kek, kek3)

	// Different salt => different KEK
	kek4 := svc.DeriveKEK("password123", []byte("differentsalt123"))
	assert.NotEqual(t, kek, kek4)
}

func TestEncryptDecryptDEK(t *testing.T) {
	svc := NewService()

	dek, err := svc.GenerateDEK()
	require.NoError(t, err)

	kek := svc.DeriveKEK("testpassword", []byte("testsalt12345678"))

	encrypted, err := svc.EncryptDEK(dek, kek)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := svc.DecryptDEK(encrypted, kek)
	require.NoError(t, err)
	assert.Equal(t, dek, decrypted)

	// Wrong KEK should fail
	wrongKEK := svc.DeriveKEK("wrongpassword", []byte("testsalt12345678"))
	_, err = svc.DecryptDEK(encrypted, wrongKEK)
	assert.Error(t, err)
}

func TestEncryptDecryptDEK_InvalidBase64(t *testing.T) {
	svc := NewService()
	kek := svc.DeriveKEK("test", []byte("testsalt12345678"))

	_, err := svc.DecryptDEK("not-valid-base64!!!", kek)
	assert.Error(t, err)
}

func TestStoreDEKAndGetDEK(t *testing.T) {
	svc := NewService()
	userID := uuid.New()
	dek := []byte("01234567890123456789012345678901")

	// Before storing, GetDEK should fail
	_, err := svc.GetDEK(userID)
	assert.Error(t, err)

	// Store and retrieve
	svc.StoreDEK(userID, dek)
	got, err := svc.GetDEK(userID)
	require.NoError(t, err)
	assert.Equal(t, dek, got)
}

func TestRemoveDEK(t *testing.T) {
	svc := NewService()
	userID := uuid.New()
	dek := []byte("01234567890123456789012345678901")

	svc.StoreDEK(userID, dek)
	_, err := svc.GetDEK(userID)
	require.NoError(t, err)

	svc.RemoveDEK(userID)
	_, err = svc.GetDEK(userID)
	assert.Error(t, err)
}

func TestEncryptDecryptField(t *testing.T) {
	svc := NewService()
	userID := uuid.New()
	dek := []byte("01234567890123456789012345678901")
	svc.StoreDEK(userID, dek)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "Hello World"},
		{"number", "12345.67"},
		{"empty string", ""},
		{"unicode", "Привет мир 日本語"},
		{"special chars", "test@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := svc.EncryptField(userID, tt.plaintext)
			require.NoError(t, err)

			if tt.plaintext == "" {
				assert.Equal(t, "", encrypted)
				return
			}

			assert.NotEqual(t, tt.plaintext, encrypted)

			decrypted, err := svc.DecryptField(userID, encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptField_NoDEK(t *testing.T) {
	svc := NewService()
	userID := uuid.New()

	_, err := svc.EncryptField(userID, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session expired")
}

func TestDecryptField_InvalidCiphertext(t *testing.T) {
	svc := NewService()
	userID := uuid.New()
	dek := []byte("01234567890123456789012345678901")
	svc.StoreDEK(userID, dek)

	_, err := svc.DecryptField(userID, "not-valid-base64!!!")
	assert.Error(t, err)
}

func TestE2EUserPassthrough(t *testing.T) {
	svc := NewService()
	userID := uuid.New()
	dek := []byte("01234567890123456789012345678901")
	svc.StoreDEK(userID, dek)

	assert.False(t, svc.IsE2EUser(userID))

	svc.SetE2EUser(userID, true)
	assert.True(t, svc.IsE2EUser(userID))

	// E2E user: EncryptField returns plaintext as-is
	result, err := svc.EncryptField(userID, "plain data")
	require.NoError(t, err)
	assert.Equal(t, "plain data", result)

	// E2E user: DecryptField returns ciphertext as-is
	result, err = svc.DecryptField(userID, "some encrypted blob")
	require.NoError(t, err)
	assert.Equal(t, "some encrypted blob", result)

	svc.SetE2EUser(userID, false)
	assert.False(t, svc.IsE2EUser(userID))
}

func TestSessionKey(t *testing.T) {
	svc := NewService()

	// Before SetSessionKey, operations should fail
	_, err := svc.EncryptDEKForSession([]byte("test"))
	assert.Error(t, err)
	_, err = svc.DecryptDEKFromSession("test")
	assert.Error(t, err)

	svc.SetSessionKey("my-jwt-secret")

	dek := []byte("01234567890123456789012345678901")
	encrypted, err := svc.EncryptDEKForSession(dek)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := svc.DecryptDEKFromSession(encrypted)
	require.NoError(t, err)
	assert.Equal(t, dek, decrypted)

	// Different session key should fail
	svc2 := NewService()
	svc2.SetSessionKey("different-secret")
	_, err = svc2.DecryptDEKFromSession(encrypted)
	assert.Error(t, err)
}

func TestEncryptBytes_DifferentOutputs(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	plaintext := []byte("same plaintext")

	enc1, err := encryptBytes(plaintext, key)
	require.NoError(t, err)
	enc2, err := encryptBytes(plaintext, key)
	require.NoError(t, err)

	// Random nonce means different ciphertexts
	assert.NotEqual(t, enc1, enc2)

	// But both decrypt to same plaintext
	dec1, err := decryptBytes(enc1, key)
	require.NoError(t, err)
	dec2, err := decryptBytes(enc2, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, dec1)
	assert.Equal(t, plaintext, dec2)
}

func TestDecryptBytes_TooShort(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	_, err := decryptBytes([]byte("short"), key)
	assert.Error(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	svc := NewService()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userID := uuid.New()
			dek := []byte("01234567890123456789012345678901")
			svc.StoreDEK(userID, dek)
			_, _ = svc.GetDEK(userID)
			svc.SetE2EUser(userID, true)
			_ = svc.IsE2EUser(userID)
			svc.SetE2EUser(userID, false)
			svc.RemoveDEK(userID)
		}()
	}

	wg.Wait()
}
