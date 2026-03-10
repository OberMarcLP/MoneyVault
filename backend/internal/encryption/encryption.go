package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type Service struct {
	mu         sync.RWMutex
	keys       map[uuid.UUID][]byte
	e2eUsers   map[uuid.UUID]bool
	sessionKey []byte
}

func NewService() *Service {
	return &Service{
		keys:     make(map[uuid.UUID][]byte),
		e2eUsers: make(map[uuid.UUID]bool),
	}
}

func (s *Service) SetE2EUser(userID uuid.UUID, enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if enabled {
		s.e2eUsers[userID] = true
	} else {
		delete(s.e2eUsers, userID)
	}
}

func (s *Service) IsE2EUser(userID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.e2eUsers[userID]
}

// SetSessionKey derives a 32-byte server-side session key from the JWT secret.
// Used to encrypt DEKs for persistence in the database.
func (s *Service) SetSessionKey(jwtSecret string) {
	s.sessionKey = argon2.IDKey([]byte(jwtSecret), []byte("moneyvault-dek-session"), 1, 32*1024, 2, 32)
}

// EncryptDEKForSession encrypts a DEK with the server-side session key for DB storage.
func (s *Service) EncryptDEKForSession(dek []byte) (string, error) {
	if s.sessionKey == nil {
		return "", errors.New("session key not initialized")
	}
	encrypted, err := encryptBytes(dek, s.sessionKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptDEKFromSession decrypts a DEK stored in the database using the server-side session key.
func (s *Service) DecryptDEKFromSession(encryptedDEK string) ([]byte, error) {
	if s.sessionKey == nil {
		return nil, errors.New("session key not initialized")
	}
	data, err := base64.StdEncoding.DecodeString(encryptedDEK)
	if err != nil {
		return nil, err
	}
	return decryptBytes(data, s.sessionKey)
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	return salt, err
}

func (s *Service) GenerateDEK() ([]byte, error) {
	dek := make([]byte, 32)
	_, err := rand.Read(dek)
	return dek, err
}

func (s *Service) DeriveKEK(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
}

func (s *Service) EncryptDEK(dek, kek []byte) (string, error) {
	encrypted, err := encryptBytes(dek, kek)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (s *Service) DecryptDEK(encryptedDEK string, kek []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedDEK)
	if err != nil {
		return nil, err
	}
	return decryptBytes(data, kek)
}

func (s *Service) StoreDEK(userID uuid.UUID, dek []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[userID] = dek
}

func (s *Service) GetDEK(userID uuid.UUID) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dek, ok := s.keys[userID]
	if !ok {
		return nil, errors.New("session expired, please login again")
	}
	return dek, nil
}

func (s *Service) RemoveDEK(userID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keys, userID)
}

func (s *Service) EncryptField(userID uuid.UUID, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	// E2E users: client encrypts, server stores as-is
	if s.IsE2EUser(userID) {
		return plaintext, nil
	}
	dek, err := s.GetDEK(userID)
	if err != nil {
		return "", err
	}
	encrypted, err := encryptBytes([]byte(plaintext), dek)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (s *Service) DecryptField(userID uuid.UUID, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	// E2E users: client decrypts, server returns as-is
	if s.IsE2EUser(userID) {
		return ciphertext, nil
	}
	dek, err := s.GetDEK(userID)
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	plaintext, err := decryptBytes(data, dek)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func encryptBytes(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptBytes(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
