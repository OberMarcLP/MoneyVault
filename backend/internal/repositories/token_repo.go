package repositories

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type TokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) RevokeToken(tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO revoked_tokens (token_hash, expires_at) VALUES ($1, $2) ON CONFLICT (token_hash) DO NOTHING`,
		tokenHash, expiresAt,
	)
	return err
}

func (r *TokenRepository) IsRevoked(tokenHash string) bool {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM revoked_tokens WHERE token_hash = $1`, tokenHash)
	return err == nil && count > 0
}

func (r *TokenRepository) CleanupExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM revoked_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
