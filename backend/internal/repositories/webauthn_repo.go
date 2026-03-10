package repositories

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"moneyvault/internal/models"
)

type WebAuthnRepository struct {
	db *sqlx.DB
}

func NewWebAuthnRepository(db *sqlx.DB) *WebAuthnRepository {
	return &WebAuthnRepository{db: db}
}

func (r *WebAuthnRepository) Create(cred *models.WebAuthnCredential) error {
	_, err := r.db.NamedExec(`
		INSERT INTO webauthn_credentials (id, user_id, name, credential_id, public_key, attestation_type, transport, sign_count, aaguid)
		VALUES (:id, :user_id, :name, :credential_id, :public_key, :attestation_type, :transport, :sign_count, :aaguid)
	`, cred)
	return err
}

func (r *WebAuthnRepository) GetByCredentialID(credentialID []byte) (*models.WebAuthnCredential, error) {
	var cred models.WebAuthnCredential
	err := r.db.Get(&cred, `SELECT * FROM webauthn_credentials WHERE credential_id = $1`, credentialID)
	return &cred, err
}

func (r *WebAuthnRepository) ListByUserID(userID uuid.UUID) ([]models.WebAuthnCredential, error) {
	var creds []models.WebAuthnCredential
	err := r.db.Select(&creds, `SELECT * FROM webauthn_credentials WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	return creds, err
}

func (r *WebAuthnRepository) UpdateSignCount(id uuid.UUID, signCount int) error {
	_, err := r.db.Exec(`UPDATE webauthn_credentials SET sign_count = $1, last_used_at = NOW() WHERE id = $2`, signCount, id)
	return err
}

func (r *WebAuthnRepository) Delete(id, userID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM webauthn_credentials WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *WebAuthnRepository) CountByUserID(userID uuid.UUID) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM webauthn_credentials WHERE user_id = $1`, userID)
	return count, err
}
