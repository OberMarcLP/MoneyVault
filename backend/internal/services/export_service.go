package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"
	"time"

	"github.com/google/uuid"
)

type ExportService struct {
	transactionRepo *repositories.TransactionRepository
	accountRepo     *repositories.AccountRepository
	categoryRepo    *repositories.CategoryRepository
	enc             *encryption.Service
}

func NewExportService(
	transactionRepo *repositories.TransactionRepository,
	accountRepo *repositories.AccountRepository,
	categoryRepo *repositories.CategoryRepository,
	enc *encryption.Service,
) *ExportService {
	return &ExportService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		enc:             enc,
	}
}

type ExportTransactionRow struct {
	Date        string `json:"date"`
	Type        string `json:"type"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Account     string `json:"account"`
}

type ExportAccountRow struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Currency    string `json:"currency"`
	Balance     string `json:"balance"`
	Institution string `json:"institution"`
	IsActive    bool   `json:"is_active"`
}

func (s *ExportService) buildTransactionRows(userID uuid.UUID, filter models.TransactionFilter) ([]ExportTransactionRow, error) {
	// Fetch all matching transactions (no pagination)
	filter.Page = 1
	filter.PerPage = 100000
	transactions, _, err := s.transactionRepo.List(userID, filter)
	if err != nil {
		return nil, err
	}

	// Build category and account lookup maps
	categories, _ := s.categoryRepo.ListByUser(userID)
	catMap := make(map[uuid.UUID]string, len(categories))
	for _, c := range categories {
		catMap[c.ID] = c.Name
	}

	accounts, _ := s.accountRepo.ListByUser(userID)
	accMap := make(map[uuid.UUID]string, len(accounts))
	for _, a := range accounts {
		name, err := s.enc.DecryptField(userID, a.Name)
		if err == nil {
			accMap[a.ID] = name
		}
	}

	rows := make([]ExportTransactionRow, 0, len(transactions))
	for _, tx := range transactions {
		amount, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			return nil, err
		}
		desc, err := s.enc.DecryptField(userID, tx.Description)
		if err != nil {
			return nil, err
		}

		catName := ""
		if tx.CategoryID != nil {
			catName = catMap[*tx.CategoryID]
		}

		rows = append(rows, ExportTransactionRow{
			Date:        tx.Date.Format("2006-01-02"),
			Type:        string(tx.Type),
			Amount:      amount,
			Currency:    tx.Currency,
			Description: desc,
			Category:    catName,
			Account:     accMap[tx.AccountID],
		})
	}
	return rows, nil
}

func (s *ExportService) ExportTransactionsCSV(userID uuid.UUID, filter models.TransactionFilter, w io.Writer) error {
	rows, err := s.buildTransactionRows(userID, filter)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	if err := writer.Write([]string{"Date", "Type", "Amount", "Currency", "Description", "Category", "Account"}); err != nil {
		return err
	}

	for _, r := range rows {
		if err := writer.Write([]string{r.Date, r.Type, r.Amount, r.Currency, r.Description, r.Category, r.Account}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ExportService) ExportTransactionsJSON(userID uuid.UUID, filter models.TransactionFilter, w io.Writer) error {
	rows, err := s.buildTransactionRows(userID, filter)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(rows)
}

func (s *ExportService) buildAccountRows(userID uuid.UUID) ([]ExportAccountRow, error) {
	accounts, err := s.accountRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	rows := make([]ExportAccountRow, 0, len(accounts))
	for _, a := range accounts {
		name, err := s.enc.DecryptField(userID, a.Name)
		if err != nil {
			return nil, err
		}
		balance, err := s.enc.DecryptField(userID, a.Balance)
		if err != nil {
			return nil, err
		}

		inst := ""
		if a.Institution != nil {
			inst = *a.Institution
		}

		rows = append(rows, ExportAccountRow{
			Name:        name,
			Type:        string(a.Type),
			Currency:    a.Currency,
			Balance:     balance,
			Institution: inst,
			IsActive:    a.IsActive,
		})
	}
	return rows, nil
}

func (s *ExportService) ExportAccountsCSV(userID uuid.UUID, w io.Writer) error {
	rows, err := s.buildAccountRows(userID)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{"Name", "Type", "Currency", "Balance", "Institution", "Active"}); err != nil {
		return err
	}

	for _, r := range rows {
		if err := writer.Write([]string{r.Name, r.Type, r.Currency, r.Balance, r.Institution, fmt.Sprintf("%t", r.IsActive)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ExportService) ExportAccountsJSON(userID uuid.UUID, w io.Writer) error {
	rows, err := s.buildAccountRows(userID)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(rows)
}

type AllExportData struct {
	Accounts     []ExportAccountRow     `json:"accounts"`
	Transactions []ExportTransactionRow `json:"transactions"`
	ExportedAt   string                 `json:"exported_at"`
}

func (s *ExportService) ExportAllJSON(userID uuid.UUID) (*AllExportData, error) {
	accounts, err := s.buildAccountRows(userID)
	if err != nil {
		return nil, err
	}
	transactions, err := s.buildTransactionRows(userID, models.TransactionFilter{})
	if err != nil {
		return nil, err
	}
	return &AllExportData{
		Accounts:     accounts,
		Transactions: transactions,
		ExportedAt:   time.Now().Format(time.RFC3339),
	}, nil
}
