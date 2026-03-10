package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionService struct {
	repo    *repositories.TransactionRepository
	accRepo *repositories.AccountRepository
	enc     *encryption.Service
}

func NewTransactionService(
	repo *repositories.TransactionRepository,
	accRepo *repositories.AccountRepository,
	enc *encryption.Service,
) *TransactionService {
	return &TransactionService{repo: repo, accRepo: accRepo, enc: enc}
}

func validateAmount(value string) error {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return fmt.Errorf("amount must be a valid number")
	}
	if d.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be greater than zero")
	}
	if d.Exponent() < -2 {
		return fmt.Errorf("amount cannot have more than 2 decimal places")
	}
	return nil
}

func (s *TransactionService) Create(userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error) {
	if err := validateAmount(req.Amount); err != nil {
		return nil, err
	}

	if _, err := s.accRepo.GetByID(req.AccountID, userID); err != nil {
		return nil, errors.New("account not found")
	}

	if req.Type == models.TransactionTransfer && req.TransferAccountID == nil {
		return nil, errors.New("transfer_account_id is required for transfer transactions")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	encAmount, err := s.enc.EncryptField(userID, req.Amount)
	if err != nil {
		return nil, err
	}
	encDesc, err := s.enc.EncryptField(userID, req.Description)
	if err != nil {
		return nil, err
	}

	tags := req.Tags
	if tags == nil {
		tags = json.RawMessage("[]")
	}

	tx := &models.Transaction{
		ID:                uuid.New(),
		AccountID:         req.AccountID,
		UserID:            userID,
		Type:              req.Type,
		Amount:            encAmount,
		Currency:          req.Currency,
		CategoryID:        req.CategoryID,
		Description:       encDesc,
		Date:              date,
		Tags:              tags,
		ImportSource:      "manual",
		TransferAccountID: req.TransferAccountID,
	}

	if err := s.repo.Create(tx); err != nil {
		return nil, err
	}

	tx.Amount = req.Amount
	tx.Description = req.Description
	return tx, nil
}

func (s *TransactionService) GetByID(id, userID uuid.UUID) (*models.Transaction, error) {
	tx, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	return s.decryptTransaction(userID, tx)
}

func (s *TransactionService) List(userID uuid.UUID, filter models.TransactionFilter) (*models.PaginatedResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 50
	}

	// When search is active, fetch all matching transactions for post-decryption filtering.
	if filter.Search != nil && *filter.Search != "" {
		searchFilter := filter
		searchFilter.Page = 1
		searchFilter.PerPage = 10000
		allTx, _, err := s.repo.List(userID, searchFilter)
		if err != nil {
			return nil, err
		}

		var filtered []models.Transaction
		needle := strings.ToLower(*filter.Search)
		for i := range allTx {
			dec, err := s.decryptTransaction(userID, &allTx[i])
			if err != nil {
				return nil, err
			}
			if strings.Contains(strings.ToLower(dec.Description), needle) {
				filtered = append(filtered, *dec)
			}
		}

		total := len(filtered)
		totalPages := total / filter.PerPage
		if total%filter.PerPage != 0 {
			totalPages++
		}

		start := (filter.Page - 1) * filter.PerPage
		end := start + filter.PerPage
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		return &models.PaginatedResponse{
			Data:       filtered[start:end],
			Total:      total,
			Page:       filter.Page,
			PerPage:    filter.PerPage,
			TotalPages: totalPages,
		}, nil
	}

	transactions, total, err := s.repo.List(userID, filter)
	if err != nil {
		return nil, err
	}

	for i := range transactions {
		decrypted, err := s.decryptTransaction(userID, &transactions[i])
		if err != nil {
			return nil, err
		}
		transactions[i] = *decrypted
	}

	totalPages := total / filter.PerPage
	if total%filter.PerPage != 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       transactions,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (s *TransactionService) Update(id, userID uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error) {
	tx, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	if req.AccountID != nil {
		tx.AccountID = *req.AccountID
	}
	if req.Type != nil {
		tx.Type = *req.Type
	}
	if req.Amount != nil {
		if err := validateAmount(*req.Amount); err != nil {
			return nil, err
		}
		encAmount, err := s.enc.EncryptField(userID, *req.Amount)
		if err != nil {
			return nil, err
		}
		tx.Amount = encAmount
	}
	if req.Currency != nil {
		tx.Currency = *req.Currency
	}
	if req.CategoryID != nil {
		tx.CategoryID = req.CategoryID
	}
	if req.Description != nil {
		encDesc, err := s.enc.EncryptField(userID, *req.Description)
		if err != nil {
			return nil, err
		}
		tx.Description = encDesc
	}
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, errors.New("invalid date format")
		}
		tx.Date = date
	}
	if req.Tags != nil {
		tx.Tags = req.Tags
	}
	if req.TransferAccountID != nil {
		tx.TransferAccountID = req.TransferAccountID
	}

	if err := s.repo.Update(tx); err != nil {
		return nil, err
	}

	return s.decryptTransaction(userID, tx)
}

func (s *TransactionService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

func (s *TransactionService) decryptTransaction(userID uuid.UUID, tx *models.Transaction) (*models.Transaction, error) {
	amount, err := s.enc.DecryptField(userID, tx.Amount)
	if err != nil {
		return nil, err
	}
	desc, err := s.enc.DecryptField(userID, tx.Description)
	if err != nil {
		return nil, err
	}
	tx.Amount = amount
	tx.Description = desc
	return tx, nil
}
