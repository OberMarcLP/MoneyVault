package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *models.Notification) error {
	query := `INSERT INTO notifications (id, user_id, type, title, message, link)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, n.ID, n.UserID, n.Type, n.Title, n.Message, n.Link)
	return err
}

func (r *NotificationRepository) List(userID uuid.UUID, unreadOnly bool, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	var err error
	if unreadOnly {
		err = r.db.Select(&notifications,
			`SELECT * FROM notifications WHERE user_id = $1 AND is_read = FALSE ORDER BY created_at DESC LIMIT $2`,
			userID, limit)
	} else {
		err = r.db.Select(&notifications,
			`SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
			userID, limit)
	}
	if notifications == nil {
		notifications = []models.Notification{}
	}
	return notifications, err
}

func (r *NotificationRepository) UnreadCount(userID uuid.UUID) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`, userID)
	return count, err
}

func (r *NotificationRepository) MarkRead(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *NotificationRepository) MarkAllRead(userID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`, userID)
	return err
}

func (r *NotificationRepository) Delete(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM notifications WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *NotificationRepository) ClearAll(userID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM notifications WHERE user_id = $1`, userID)
	return err
}

// Alert Rules
func (r *NotificationRepository) CreateRule(rule *models.AlertRule) error {
	query := `INSERT INTO alert_rules (id, user_id, type, condition, is_active)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, rule.ID, rule.UserID, rule.Type, rule.Condition, rule.IsActive)
	return err
}

func (r *NotificationRepository) ListRules(userID uuid.UUID) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.Select(&rules, `SELECT * FROM alert_rules WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`, userID)
	if rules == nil {
		rules = []models.AlertRule{}
	}
	return rules, err
}

func (r *NotificationRepository) GetActiveRules() ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.Select(&rules, `SELECT * FROM alert_rules WHERE is_active = TRUE AND deleted_at IS NULL`)
	if rules == nil {
		rules = []models.AlertRule{}
	}
	return rules, err
}

func (r *NotificationRepository) ToggleRule(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE alert_rules SET is_active = NOT is_active WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return err
}

func (r *NotificationRepository) DeleteRule(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE alert_rules SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return err
}

func (r *NotificationRepository) UpdateLastTriggered(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE alert_rules SET last_triggered = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
