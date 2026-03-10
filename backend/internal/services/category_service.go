package services

import (
	"errors"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

type CategoryService struct {
	repo *repositories.CategoryRepository
}

func NewCategoryService(repo *repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(userID uuid.UUID, req models.CreateCategoryRequest) (*models.Category, error) {
	cat := &models.Category{
		ID:       uuid.New(),
		UserID:   userID,
		Name:     req.Name,
		Type:     req.Type,
		Icon:     req.Icon,
		Color:    req.Color,
		ParentID: req.ParentID,
	}

	if err := s.repo.Create(cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) GetByID(id, userID uuid.UUID) (*models.Category, error) {
	return s.repo.GetByID(id, userID)
}

func (s *CategoryService) List(userID uuid.UUID) ([]models.Category, error) {
	return s.repo.ListByUser(userID)
}

func (s *CategoryService) Update(id, userID uuid.UUID, req models.UpdateCategoryRequest) (*models.Category, error) {
	cat, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.Type != nil {
		cat.Type = *req.Type
	}
	if req.Icon != nil {
		cat.Icon = *req.Icon
	}
	if req.Color != nil {
		cat.Color = *req.Color
	}
	if req.ParentID != nil {
		cat.ParentID = req.ParentID
	}

	if err := s.repo.Update(cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}
