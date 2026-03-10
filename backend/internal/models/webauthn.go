package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type WebAuthnCredential struct {
	ID              uuid.UUID      `db:"id" json:"id"`
	UserID          uuid.UUID      `db:"user_id" json:"user_id"`
	Name            string         `db:"name" json:"name"`
	CredentialID    []byte         `db:"credential_id" json:"-"`
	PublicKey       []byte         `db:"public_key" json:"-"`
	AttestationType string         `db:"attestation_type" json:"-"`
	Transport       pq.StringArray `db:"transport" json:"-"`
	SignCount       int            `db:"sign_count" json:"-"`
	AAGUID          []byte         `db:"aaguid" json:"-"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	LastUsedAt      *time.Time     `db:"last_used_at" json:"last_used_at"`
}
