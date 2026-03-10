package services

import (
	"errors"
	"fmt"
	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountService struct {
	repo *repositories.AccountRepository
	enc  *encryption.Service
}

func NewAccountService(repo *repositories.AccountRepository, enc *encryption.Service) *AccountService {
	return &AccountService{repo: repo, enc: enc}
}

func validateDecimal(value, field string) error {
	if value == "" {
		return nil
	}
	d, err := decimal.NewFromString(value)
	if err != nil {
		return fmt.Errorf("%s must be a valid number", field)
	}
	if d.Exponent() < -2 {
		return fmt.Errorf("%s cannot have more than 2 decimal places", field)
	}
	return nil
}

func (s *AccountService) Create(userID uuid.UUID, req models.CreateAccountRequest) (*models.Account, error) {
	if err := validateDecimal(req.Balance, "balance"); err != nil {
		return nil, err
	}

	encName, err := s.enc.EncryptField(userID, req.Name)
	if err != nil {
		return nil, err
	}
	encBalance, err := s.enc.EncryptField(userID, req.Balance)
	if err != nil {
		return nil, err
	}

	account := &models.Account{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        encName,
		Type:        req.Type,
		Currency:    req.Currency,
		Balance:     encBalance,
		Institution: req.Institution,
		IsActive:    true,
	}

	if err := s.repo.Create(account); err != nil {
		return nil, err
	}

	account.Name = req.Name
	account.Balance = req.Balance
	return account, nil
}

func (s *AccountService) GetByID(id, userID uuid.UUID) (*models.Account, error) {
	account, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	return s.decryptAccount(userID, account)
}

func (s *AccountService) List(userID uuid.UUID) ([]models.Account, error) {
	accounts, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	for i := range accounts {
		decrypted, err := s.decryptAccount(userID, &accounts[i])
		if err != nil {
			return nil, err
		}
		accounts[i] = *decrypted
	}
	return accounts, nil
}

func (s *AccountService) Update(id, userID uuid.UUID, req models.UpdateAccountRequest) (*models.Account, error) {
	account, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, errors.New("account not found")
	}

	if req.Name != nil {
		encName, err := s.enc.EncryptField(userID, *req.Name)
		if err != nil {
			return nil, err
		}
		account.Name = encName
	}
	if req.Type != nil {
		account.Type = *req.Type
	}
	if req.Currency != nil {
		account.Currency = *req.Currency
	}
	if req.Balance != nil {
		if err := validateDecimal(*req.Balance, "balance"); err != nil {
			return nil, err
		}
		encBalance, err := s.enc.EncryptField(userID, *req.Balance)
		if err != nil {
			return nil, err
		}
		account.Balance = encBalance
	}
	if req.Institution != nil {
		account.Institution = req.Institution
	}
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	if err := s.repo.Update(account); err != nil {
		return nil, err
	}

	return s.decryptAccount(userID, account)
}

func (s *AccountService) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

func (s *AccountService) decryptAccount(userID uuid.UUID, account *models.Account) (*models.Account, error) {
	name, err := s.enc.DecryptField(userID, account.Name)
	if err != nil {
		return nil, err
	}
	balance, err := s.enc.DecryptField(userID, account.Balance)
	if err != nil {
		return nil, err
	}
	account.Name = name
	account.Balance = balance
	return account, nil
}
