package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(cat *models.Category) error {
	query := `
		INSERT INTO categories (id, user_id, name, type, icon, color, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`
	return r.db.QueryRow(query,
		cat.ID, cat.UserID, cat.Name, cat.Type,
		cat.Icon, cat.Color, cat.ParentID,
	).Scan(&cat.CreatedAt)
}

// CreateWithTx creates a category within a database transaction.
func (r *CategoryRepository) CreateWithTx(dbTx *sqlx.Tx, cat *models.Category) error {
	query := `
		INSERT INTO categories (id, user_id, name, type, icon, color, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`
	return dbTx.QueryRow(query,
		cat.ID, cat.UserID, cat.Name, cat.Type,
		cat.Icon, cat.Color, cat.ParentID,
	).Scan(&cat.CreatedAt)
}

func (r *CategoryRepository) GetByID(id, userID uuid.UUID) (*models.Category, error) {
	var cat models.Category
	err := r.db.Get(&cat, "SELECT * FROM categories WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepository) ListByUser(userID uuid.UUID) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Select(&categories, "SELECT * FROM categories WHERE user_id = $1 ORDER BY type, name", userID)
	return categories, err
}

func (r *CategoryRepository) Update(cat *models.Category) error {
	query := `
		UPDATE categories SET name = $3, type = $4, icon = $5,
		color = $6, parent_id = $7
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(query,
		cat.ID, cat.UserID, cat.Name, cat.Type,
		cat.Icon, cat.Color, cat.ParentID,
	)
	return err
}

func (r *CategoryRepository) Delete(id, userID uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM categories WHERE id = $1 AND user_id = $2", id, userID)
	return err
}

func (r *CategoryRepository) CreateDefaults(userID uuid.UUID) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO categories (id, user_id, name, type, icon, color, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())`
	for _, cat := range models.DefaultCategories {
		_, err := tx.Exec(query, uuid.New(), userID, cat.Name, cat.Type, cat.Icon, cat.Color)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
