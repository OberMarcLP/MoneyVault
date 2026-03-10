package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ExchangeConnectionRepository struct {
	db *sqlx.DB
}

func NewExchangeConnectionRepository(db *sqlx.DB) *ExchangeConnectionRepository {
	return &ExchangeConnectionRepository{db: db}
}

func (r *ExchangeConnectionRepository) Create(conn *models.ExchangeConnection) error {
	query := `INSERT INTO exchange_connections (id, user_id, exchange, api_key, api_secret, label, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`
	return r.db.QueryRow(query, conn.ID, conn.UserID, conn.Exchange, conn.APIKey, conn.APISecret, conn.Label, conn.IsActive).Scan(&conn.CreatedAt, &conn.UpdatedAt)
}

func (r *ExchangeConnectionRepository) GetByID(userID, id uuid.UUID) (*models.ExchangeConnection, error) {
	var conn models.ExchangeConnection
	err := r.db.Get(&conn, `SELECT * FROM exchange_connections WHERE id = $1 AND user_id = $2`, id, userID)
	return &conn, err
}

func (r *ExchangeConnectionRepository) List(userID uuid.UUID) ([]models.ExchangeConnection, error) {
	var conns []models.ExchangeConnection
	err := r.db.Select(&conns, `SELECT * FROM exchange_connections WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if conns == nil {
		conns = []models.ExchangeConnection{}
	}
	return conns, err
}

func (r *ExchangeConnectionRepository) Delete(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM exchange_connections WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *ExchangeConnectionRepository) UpdateLastSynced(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE exchange_connections SET last_synced = NOW(), updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *ExchangeConnectionRepository) SetActive(userID, id uuid.UUID, active bool) error {
	_, err := r.db.Exec(`UPDATE exchange_connections SET is_active = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`, active, id, userID)
	return err
}
