package services

import (
	"fmt"
	"time"

	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

type RecurringService struct {
	recurringRepo   *repositories.RecurringRepository
	transactionRepo *repositories.TransactionRepository
	accountRepo     *repositories.AccountRepository
	enc             *encryption.Service
}

func NewRecurringService(
	recurringRepo *repositories.RecurringRepository,
	transactionRepo *repositories.TransactionRepository,
	accountRepo *repositories.AccountRepository,
	enc *encryption.Service,
) *RecurringService {
	return &RecurringService{
		recurringRepo:   recurringRepo,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		enc:             enc,
	}
}

func (s *RecurringService) Create(userID uuid.UUID, req models.CreateRecurringRequest) (*models.RecurringTransaction, error) {
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID")
	}

	nextDate, err := time.Parse("2006-01-02", req.NextDate)
	if err != nil {
		return nil, fmt.Errorf("invalid next_date format")
	}

	rt := &models.RecurringTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		AccountID:   accountID,
		Type:        req.Type,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Frequency:   models.Frequency(req.Frequency),
		NextDate:    nextDate,
		IsActive:    true,
	}

	if rt.Currency == "" {
		rt.Currency = "USD"
	}

	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category ID")
		}
		rt.CategoryID = &catID
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format")
		}
		rt.EndDate = &endDate
	}

	if req.TransferAccountID != "" {
		taID, err := uuid.Parse(req.TransferAccountID)
		if err != nil {
			return nil, fmt.Errorf("invalid transfer_account_id")
		}
		rt.TransferAccountID = &taID
	}

	if err := s.recurringRepo.Create(rt); err != nil {
		return nil, fmt.Errorf("failed to create recurring transaction: %w", err)
	}

	return rt, nil
}

func (s *RecurringService) List(userID uuid.UUID) ([]models.RecurringTransaction, error) {
	return s.recurringRepo.List(userID)
}

func (s *RecurringService) GetByID(userID, id uuid.UUID) (*models.RecurringTransaction, error) {
	return s.recurringRepo.GetByID(userID, id)
}

func (s *RecurringService) Delete(userID, id uuid.UUID) error {
	return s.recurringRepo.Delete(userID, id)
}

func (s *RecurringService) ToggleActive(userID, id uuid.UUID) (*models.RecurringTransaction, error) {
	rt, err := s.recurringRepo.GetByID(userID, id)
	if err != nil {
		return nil, err
	}
	rt.IsActive = !rt.IsActive
	if err := s.recurringRepo.Update(rt); err != nil {
		return nil, err
	}
	return rt, nil
}

func (s *RecurringService) ProcessDue() (int, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	due, err := s.recurringRepo.GetDue(today.Add(24 * time.Hour))
	if err != nil {
		return 0, err
	}

	created := 0
	for _, rt := range due {
		if _, err := s.enc.GetDEK(rt.UserID); err != nil {
			continue
		}

		encAmount, err := s.enc.EncryptField(rt.UserID, rt.Amount)
		if err != nil {
			continue
		}
		encDesc, err := s.enc.EncryptField(rt.UserID, rt.Description)
		if err != nil {
			continue
		}

		tx := &models.Transaction{
			ID:                uuid.New(),
			AccountID:         rt.AccountID,
			UserID:            rt.UserID,
			Type:              models.TransactionType(rt.Type),
			Amount:            encAmount,
			Currency:          rt.Currency,
			CategoryID:        rt.CategoryID,
			Description:       encDesc,
			Date:              rt.NextDate,
			ImportSource:      "recurring",
			TransferAccountID: rt.TransferAccountID,
		}

		if err := s.transactionRepo.Create(tx); err != nil {
			continue
		}

		now := time.Now()
		rt.LastCreated = &now
		rt.NextDate = advanceDate(rt.NextDate, rt.Frequency)

		if rt.EndDate != nil && rt.NextDate.After(*rt.EndDate) {
			rt.IsActive = false
		}

		_ = s.recurringRepo.Update(&rt)
		created++
	}

	return created, nil
}

func advanceDate(from time.Time, freq models.Frequency) time.Time {
	switch freq {
	case models.FreqDaily:
		return from.AddDate(0, 0, 1)
	case models.FreqWeekly:
		return from.AddDate(0, 0, 7)
	case models.FreqBiweekly:
		return from.AddDate(0, 0, 14)
	case models.FreqMonthly:
		return from.AddDate(0, 1, 0)
	case models.FreqQuarterly:
		return from.AddDate(0, 3, 0)
	case models.FreqYearly:
		return from.AddDate(1, 0, 0)
	default:
		return from.AddDate(0, 1, 0)
	}
}
