package services

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

type BudgetService struct {
	budgetRepo      *repositories.BudgetRepository
	categoryRepo    *repositories.CategoryRepository
	transactionRepo *repositories.TransactionRepository
	enc             *encryption.Service
}

func NewBudgetService(budgetRepo *repositories.BudgetRepository, categoryRepo *repositories.CategoryRepository, transactionRepo *repositories.TransactionRepository, enc *encryption.Service) *BudgetService {
	return &BudgetService{budgetRepo: budgetRepo, categoryRepo: categoryRepo, transactionRepo: transactionRepo, enc: enc}
}

func (s *BudgetService) Create(userID uuid.UUID, req models.CreateBudgetRequest) (*models.Budget, error) {
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format")
	}

	budget := &models.Budget{
		ID:         uuid.New(),
		UserID:     userID,
		CategoryID: &categoryID,
		Amount:     req.Amount,
		Period:     models.BudgetPeriod(req.Period),
		StartDate:  startDate,
		Rollover:   req.Rollover,
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format")
		}
		budget.EndDate = &endDate
	}

	if err := s.budgetRepo.Create(budget); err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	return budget, nil
}

func (s *BudgetService) List(userID uuid.UUID) ([]models.BudgetWithSpending, error) {
	budgets, err := s.budgetRepo.List(userID)
	if err != nil {
		return nil, err
	}

	categories, err := s.categoryRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	catMap := make(map[uuid.UUID]models.Category)
	for _, c := range categories {
		catMap[c.ID] = c
	}

	var categoryIDs []uuid.UUID
	for _, b := range budgets {
		if b.CategoryID != nil {
			categoryIDs = append(categoryIDs, *b.CategoryID)
		}
	}

	now := time.Now()
	var results []models.BudgetWithSpending

	spentMap := make(map[uuid.UUID]float64)
	if len(categoryIDs) > 0 {
		for _, b := range budgets {
			if b.CategoryID == nil {
				continue
			}
			start, end := periodBounds(now, b.Period)
			spent, err := s.calcSpentForCategory(userID, b.CategoryID, start, end)
			if err != nil {
				continue
			}
			spentMap[b.ID] = spent
		}
	}

	for _, b := range budgets {
		bws := models.BudgetWithSpending{Budget: b}

		if b.CategoryID != nil {
			if cat, ok := catMap[*b.CategoryID]; ok {
				bws.CategoryName = cat.Name
				bws.CategoryIcon = cat.Icon
				bws.CategoryColor = cat.Color
			}
		}

		bws.Spent = spentMap[b.ID]
		bws.Remaining = math.Max(0, b.Amount-bws.Spent)
		if b.Amount > 0 {
			bws.Percentage = math.Min(100, (bws.Spent/b.Amount)*100)
		}

		results = append(results, bws)
	}

	if results == nil {
		results = []models.BudgetWithSpending{}
	}
	return results, nil
}

func (s *BudgetService) GetByID(userID, budgetID uuid.UUID) (*models.BudgetWithSpending, error) {
	budget, err := s.budgetRepo.GetByID(userID, budgetID)
	if err != nil {
		return nil, err
	}

	bws := &models.BudgetWithSpending{Budget: *budget}

	categories, err := s.categoryRepo.ListByUser(userID)
	if err == nil {
		for _, c := range categories {
			if budget.CategoryID != nil && c.ID == *budget.CategoryID {
				bws.CategoryName = c.Name
				bws.CategoryIcon = c.Icon
				bws.CategoryColor = c.Color
				break
			}
		}
	}

	if budget.CategoryID != nil {
		now := time.Now()
		start, end := periodBounds(now, budget.Period)
		bws.Spent, _ = s.calcSpentForCategory(userID, budget.CategoryID, start, end)
		bws.Remaining = math.Max(0, budget.Amount-bws.Spent)
		if budget.Amount > 0 {
			bws.Percentage = math.Min(100, (bws.Spent/budget.Amount)*100)
		}
	}

	return bws, nil
}

func (s *BudgetService) Update(userID, budgetID uuid.UUID, req models.UpdateBudgetRequest) (*models.Budget, error) {
	budget, err := s.budgetRepo.GetByID(userID, budgetID)
	if err != nil {
		return nil, err
	}

	if req.Amount != nil {
		budget.Amount = *req.Amount
	}
	if req.Period != "" {
		budget.Period = models.BudgetPeriod(req.Period)
	}
	if req.Rollover != nil {
		budget.Rollover = *req.Rollover
	}
	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format")
		}
		budget.EndDate = &endDate
	}

	if err := s.budgetRepo.Update(budget); err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *BudgetService) Delete(userID, budgetID uuid.UUID) error {
	return s.budgetRepo.Delete(userID, budgetID)
}

// calcSpentForCategory fetches encrypted transactions and sums amounts in application code.
func (s *BudgetService) calcSpentForCategory(userID uuid.UUID, categoryID *uuid.UUID, start, end time.Time) (float64, error) {
	txs, err := s.transactionRepo.ListExpensesByCategory(userID, categoryID, start, end)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, tx := range txs {
		plain, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(plain, 64)
		if err != nil {
			continue
		}
		total += amt
	}
	return total, nil
}

func periodBounds(now time.Time, period models.BudgetPeriod) (time.Time, time.Time) {
	switch period {
	case models.PeriodWeekly:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -(weekday - 1))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 0, 7)
		return start, end
	case models.PeriodYearly:
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(1, 0, 0)
		return start, end
	default: // monthly
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0)
		return start, end
	}
}
