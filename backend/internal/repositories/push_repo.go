package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PushRepository struct {
	db *sqlx.DB
}

func NewPushRepository(db *sqlx.DB) *PushRepository {
	return &PushRepository{db: db}
}

func (r *PushRepository) Subscribe(sub *models.PushSubscription) error {
	query := `INSERT INTO push_subscriptions (id, user_id, endpoint, auth, p256dh)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, endpoint) DO UPDATE SET auth = $4, p256dh = $5`
	_, err := r.db.Exec(query, sub.ID, sub.UserID, sub.Endpoint, sub.Auth, sub.P256dh)
	return err
}

func (r *PushRepository) Unsubscribe(userID uuid.UUID, endpoint string) error {
	_, err := r.db.Exec(`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`, userID, endpoint)
	return err
}

func (r *PushRepository) GetByUser(userID uuid.UUID) ([]models.PushSubscription, error) {
	var subs []models.PushSubscription
	err := r.db.Select(&subs, `SELECT * FROM push_subscriptions WHERE user_id = $1`, userID)
	if subs == nil {
		subs = []models.PushSubscription{}
	}
	return subs, err
}

func (r *PushRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM push_subscriptions WHERE id = $1`, id)
	return err
}
