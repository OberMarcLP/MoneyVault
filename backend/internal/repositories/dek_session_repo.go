package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type DEKSession struct {
	UserID       uuid.UUID `db:"user_id"`
	EncryptedDEK string    `db:"encrypted_dek"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}

type DEKSessionRepository struct {
	db *sqlx.DB
}

func NewDEKSessionRepository(db *sqlx.DB) *DEKSessionRepository {
	return &DEKSessionRepository{db: db}
}

func (r *DEKSessionRepository) Upsert(session *DEKSession) error {
	_, err := r.db.Exec(
		`INSERT INTO dek_sessions (user_id, encrypted_dek, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET encrypted_dek = $2, expires_at = $3`,
		session.UserID, session.EncryptedDEK, session.ExpiresAt,
	)
	return err
}

func (r *DEKSessionRepository) GetByUserID(userID uuid.UUID) (*DEKSession, error) {
	var session DEKSession
	err := r.db.Get(&session,
		`SELECT user_id, encrypted_dek, expires_at, created_at FROM dek_sessions
		 WHERE user_id = $1 AND expires_at > NOW()`, userID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *DEKSessionRepository) Delete(userID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM dek_sessions WHERE user_id = $1`, userID)
	return err
}

func (r *DEKSessionRepository) GetAllActive() ([]DEKSession, error) {
	var sessions []DEKSession
	err := r.db.Select(&sessions,
		`SELECT user_id, encrypted_dek, expires_at, created_at FROM dek_sessions WHERE expires_at > NOW()`)
	return sessions, err
}

func (r *DEKSessionRepository) CleanupExpired() (int64, error) {
	res, err := r.db.Exec(`DELETE FROM dek_sessions WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
