package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"moneyvault/internal/encryption"
	"moneyvault/internal/integrations"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AnalyticsService struct {
	analyticsRepo   *repositories.AnalyticsRepository
	accountRepo     *repositories.AccountRepository
	holdingRepo     *repositories.HoldingRepository
	categoryRepo    *repositories.CategoryRepository
	budgetRepo      *repositories.BudgetRepository
	recurringRepo   *repositories.RecurringRepository
	transactionRepo *repositories.TransactionRepository
	enc             *encryption.Service
	exchangeRate    *integrations.ExchangeRateClient
}

func NewAnalyticsService(
	analyticsRepo *repositories.AnalyticsRepository,
	accountRepo *repositories.AccountRepository,
	holdingRepo *repositories.HoldingRepository,
	categoryRepo *repositories.CategoryRepository,
	budgetRepo *repositories.BudgetRepository,
	recurringRepo *repositories.RecurringRepository,
	transactionRepo *repositories.TransactionRepository,
	enc *encryption.Service,
	exchangeRate *integrations.ExchangeRateClient,
) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo:   analyticsRepo,
		accountRepo:     accountRepo,
		holdingRepo:     holdingRepo,
		categoryRepo:    categoryRepo,
		budgetRepo:      budgetRepo,
		recurringRepo:   recurringRepo,
		transactionRepo: transactionRepo,
		enc:             enc,
		exchangeRate:    exchangeRate,
	}
}

// decryptAccountBalances reads raw rows and decrypts the balance for each,
// converting all values to the target currency using exchange rates.
func (s *AnalyticsService) decryptAccountBalances(userID uuid.UUID, targetCurrency string) (float64, []byte, error) {
	rows, err := s.analyticsRepo.GetAccountBalances(userID)
	if err != nil {
		return 0, nil, err
	}

	total := decimal.Zero
	breakdown := make(map[string]decimal.Decimal)
	for _, row := range rows {
		balStr, dErr := s.enc.DecryptField(userID, row.Balance)
		if dErr != nil {
			continue
		}
		bal, pErr := decimal.NewFromString(balStr)
		if pErr != nil {
			continue
		}

		// Convert to target currency if exchange rates are available
		if row.Currency != targetCurrency && s.exchangeRate != nil && s.exchangeRate.IsLoaded() {
			rate, rErr := s.exchangeRate.GetRate(row.Currency, targetCurrency)
			if rErr == nil {
				bal = bal.Mul(decimal.NewFromFloat(rate))
			}
		}

		total = total.Add(bal)
		breakdown[row.Type] = breakdown[row.Type].Add(bal)
	}

	floatBreakdown := make(map[string]float64)
	for k, v := range breakdown {
		floatBreakdown[k], _ = v.Float64()
	}
	bJSON, _ := json.Marshal(floatBreakdown)
	f, _ := total.Float64()
	return f, bJSON, nil
}

// TakeSnapshot creates a net worth snapshot for a user (stored in USD)
func (s *AnalyticsService) TakeSnapshot(userID uuid.UUID) error {
	accountsValue, breakdownJSON, err := s.decryptAccountBalances(userID, "USD")
	if err != nil {
		return err
	}

	holdings, err := s.holdingRepo.List(userID)
	if err != nil {
		return err
	}

	var investmentsValue, cryptoValue float64
	var symbols []string
	for _, h := range holdings {
		symbols = append(symbols, h.Symbol)
	}
	prices, _ := s.holdingRepo.GetPrices(symbols)

	for _, h := range holdings {
		value := h.CostBasis
		if p, ok := prices[h.Symbol]; ok {
			value = p.Price * h.Quantity
		}
		if h.AssetType == models.AssetCrypto || h.AssetType == models.AssetDeFi {
			cryptoValue += value
		} else {
			investmentsValue += value
		}
	}

	today := time.Now()
	snapshot := &models.NetWorthSnapshot{
		ID:               uuid.New(),
		UserID:           userID,
		Date:             time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()),
		TotalValue:       accountsValue + investmentsValue + cryptoValue,
		AccountsValue:    accountsValue,
		InvestmentsValue: investmentsValue,
		CryptoValue:      cryptoValue,
		Breakdown:        breakdownJSON,
	}

	return s.analyticsRepo.UpsertNetWorthSnapshot(snapshot)
}

// TakeAllSnapshots runs the snapshot for every user (background job)
func (s *AnalyticsService) TakeAllSnapshots() {
	userIDs, err := s.analyticsRepo.GetAllUserIDs()
	if err != nil {
		log.Printf("Analytics snapshot: failed to get user IDs: %v", err)
		return
	}
	for _, uid := range userIDs {
		if err := s.TakeSnapshot(uid); err != nil {
			log.Printf("Analytics snapshot error for user %s: %v", uid, err)
		}
	}
}

// GetNetWorthHistory returns net worth over time
func (s *AnalyticsService) GetNetWorthHistory(userID uuid.UUID, days int) ([]models.NetWorthSnapshot, error) {
	if days <= 0 {
		days = 90
	}
	return s.analyticsRepo.GetNetWorthHistory(userID, days)
}

// GetSpendingBreakdown returns spending grouped by category (decrypts amounts in app code)
func (s *AnalyticsService) GetSpendingBreakdown(userID uuid.UUID, period string) ([]models.SpendingByCategory, error) {
	from, to := parsePeriod(period)

	txs, err := s.transactionRepo.ListExpensesByDateRange(userID, from, to)
	if err != nil {
		return nil, err
	}

	categories, _ := s.categoryRepo.ListByUser(userID)
	catMap := make(map[string]models.Category)
	for _, c := range categories {
		catMap[c.ID.String()] = c
	}

	// Decrypt and aggregate by category
	type catAgg struct {
		total float64
		count int
	}
	aggMap := make(map[string]*catAgg)
	for _, tx := range txs {
		if tx.CategoryID == nil {
			continue
		}
		catID := tx.CategoryID.String()
		plain, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(plain, 64)
		if err != nil {
			continue
		}
		if _, ok := aggMap[catID]; !ok {
			aggMap[catID] = &catAgg{}
		}
		aggMap[catID].total += amt
		aggMap[catID].count++
	}

	var result []models.SpendingByCategory
	for catID, agg := range aggMap {
		item := models.SpendingByCategory{
			CategoryID: catID,
			Total:      agg.total,
			Count:      agg.count,
		}
		if cat, ok := catMap[catID]; ok {
			item.CategoryName = cat.Name
			item.CategoryColor = cat.Color
			item.CategoryIcon = cat.Icon
		}
		result = append(result, item)
	}
	// Sort by total descending
	sort.Slice(result, func(i, j int) bool { return result[i].Total > result[j].Total })
	if result == nil {
		result = []models.SpendingByCategory{}
	}
	return result, nil
}

// GetSpendingTrends returns monthly income vs expense (decrypts amounts in app code)
func (s *AnalyticsService) GetSpendingTrends(userID uuid.UUID, months int) ([]models.SpendingTrend, error) {
	if months <= 0 {
		months = 12
	}
	from := time.Now().AddDate(0, -months, 0)
	to := time.Now().AddDate(0, 0, 1) // tomorrow to include today

	txs, err := s.transactionRepo.ListByDateRange(userID, from, to)
	if err != nil {
		return nil, err
	}

	trendMap := make(map[string]*models.SpendingTrend)
	for _, tx := range txs {
		if tx.Type != models.TransactionIncome && tx.Type != models.TransactionExpense {
			continue
		}
		plain, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(plain, 64)
		if err != nil {
			continue
		}
		period := formatPeriod(tx.Date)
		if _, ok := trendMap[period]; !ok {
			trendMap[period] = &models.SpendingTrend{Period: period}
		}
		if tx.Type == models.TransactionIncome {
			trendMap[period].Income += amt
		} else {
			trendMap[period].Expense += amt
		}
	}

	var result []models.SpendingTrend
	for _, t := range trendMap {
		t.Net = t.Income - t.Expense
		result = append(result, *t)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Period < result[j].Period })
	if result == nil {
		result = []models.SpendingTrend{}
	}
	return result, nil
}

// GetTopExpenses returns the largest expenses in a period (decrypts amounts in app code)
func (s *AnalyticsService) GetTopExpenses(userID uuid.UUID, period string, limit int) ([]models.TopExpense, error) {
	if limit <= 0 {
		limit = 10
	}
	from, to := parsePeriod(period)

	txs, err := s.transactionRepo.ListExpensesByDateRange(userID, from, to)
	if err != nil {
		return nil, err
	}

	categories, _ := s.categoryRepo.ListByUser(userID)
	catMap := make(map[string]string)
	for _, c := range categories {
		catMap[c.ID.String()] = c.Name
	}

	var result []models.TopExpense
	for _, tx := range txs {
		plainAmt, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(plainAmt, 64)
		if err != nil {
			continue
		}
		plainDesc, _ := s.enc.DecryptField(userID, tx.Description)
		if plainDesc == "" {
			plainDesc = tx.Description
		}
		te := models.TopExpense{
			Description: plainDesc,
			Amount:      amt,
			Date:        tx.Date.Format("2006-01-02"),
		}
		if tx.CategoryID != nil {
			if name, ok := catMap[tx.CategoryID.String()]; ok {
				te.Category = name
			}
		}
		result = append(result, te)
	}

	// Sort by amount descending and take top N
	sort.Slice(result, func(i, j int) bool { return result[i].Amount > result[j].Amount })
	if len(result) > limit {
		result = result[:limit]
	}
	if result == nil {
		result = []models.TopExpense{}
	}
	return result, nil
}

// GetBudgetVsActual returns budget comparison for a given period (decrypts amounts in app code)
func (s *AnalyticsService) GetBudgetVsActual(userID uuid.UUID, period string) (*models.BudgetHistory, error) {
	from, to := parsePeriod(period)

	budgets, err := s.budgetRepo.List(userID)
	if err != nil {
		return nil, err
	}

	categories, _ := s.categoryRepo.ListByUser(userID)
	catMap := make(map[string]models.Category)
	for _, c := range categories {
		catMap[c.ID.String()] = c
	}

	// Fetch all expense transactions in the period and decrypt amounts
	txs, err := s.transactionRepo.ListExpensesByDateRange(userID, from, to)
	if err != nil {
		return nil, err
	}

	// Build spending map by category
	spending := make(map[uuid.UUID]float64)
	for _, tx := range txs {
		if tx.CategoryID == nil {
			continue
		}
		plain, dErr := s.enc.DecryptField(userID, tx.Amount)
		if dErr != nil {
			continue
		}
		amt, pErr := strconv.ParseFloat(plain, 64)
		if pErr != nil {
			continue
		}
		spending[*tx.CategoryID] += amt
	}

	result := &models.BudgetHistory{
		Period: period,
	}

	for _, b := range budgets {
		if b.CategoryID == nil {
			continue
		}
		spent := spending[*b.CategoryID]

		bva := models.BudgetVsActual{
			CategoryID:   b.CategoryID.String(),
			BudgetAmount: b.Amount,
			ActualAmount: spent,
			Difference:   b.Amount - spent,
		}
		if b.Amount > 0 {
			bva.Percentage = (spent / b.Amount) * 100
		}
		if cat, ok := catMap[b.CategoryID.String()]; ok {
			bva.CategoryName = cat.Name
			bva.CategoryColor = cat.Color
		}
		result.Budgets = append(result.Budgets, bva)
		result.TotalBud += b.Amount
		result.TotalAct += spent
	}
	if result.Budgets == nil {
		result.Budgets = []models.BudgetVsActual{}
	}
	return result, nil
}

// GetCashFlowForecast projects future cash flow based on recurring transactions
func (s *AnalyticsService) GetCashFlowForecast(userID uuid.UUID, months int) (*CashFlowResult, error) {
	if months <= 0 {
		months = 6
	}

	recurring, err := s.recurringRepo.List(userID)
	if err != nil {
		return nil, err
	}

	accountsValue, _, err := s.decryptAccountBalances(userID, "USD")
	if err != nil {
		return nil, err
	}

	avgIncome, avgExpenses := s.calcMonthlyAverages(userID, 6)

	now := time.Now()
	var forecasts []models.CashFlowForecast
	runningBalance := accountsValue

	for i := 0; i < months; i++ {
		month := now.AddDate(0, i+1, 0)
		period := month.Format("2006-01")

		projIncome := avgIncome
		projExpense := avgExpenses

		for _, r := range recurring {
			if !r.IsActive {
				continue
			}
			amt, err := decimal.NewFromString(r.Amount)
			if err != nil {
				continue
			}
			freq := decimal.NewFromFloat(monthlyMultiplier(string(r.Frequency)))
			contribution, _ := amt.Mul(freq).Float64()
			if string(r.Type) == "income" {
				projIncome += contribution
			} else if string(r.Type) == "expense" {
				projExpense += contribution
			}
		}

		net := projIncome - projExpense
		runningBalance += net

		forecasts = append(forecasts, models.CashFlowForecast{
			Period:           period,
			ProjectedIncome:  math.Round(projIncome*100) / 100,
			ProjectedExpense: math.Round(projExpense*100) / 100,
			NetCashFlow:      math.Round(net*100) / 100,
			RunningBalance:   math.Round(runningBalance*100) / 100,
		})
	}

	monthlySavings := avgIncome - avgExpenses
	var runwayMonths float64
	if avgExpenses > 0 && monthlySavings < 0 {
		runwayMonths = accountsValue / avgExpenses
	} else if monthlySavings >= 0 {
		runwayMonths = -1 // infinite (savings positive)
	}

	var bills []models.UpcomingBill
	accounts, _ := s.accountRepo.ListByUser(userID)
	acctMap := make(map[string]string)
	for _, a := range accounts {
		acctMap[a.ID.String()] = a.Name
	}

	for _, r := range recurring {
		if !r.IsActive || string(r.Type) != "expense" {
			continue
		}
		if r.NextDate.Before(now.AddDate(0, 1, 0)) {
			amtDec, _ := decimal.NewFromString(r.Amount)
			amt, _ := amtDec.Float64()
			bill := models.UpcomingBill{
				Description: r.Description,
				Amount:      amt,
				DueDate:     r.NextDate.Format("2006-01-02"),
				Frequency:   string(r.Frequency),
			}
			if name, ok := acctMap[r.AccountID.String()]; ok {
				bill.AccountName = name
			}
			bills = append(bills, bill)
		}
	}
	if bills == nil {
		bills = []models.UpcomingBill{}
	}

	return &CashFlowResult{
		Forecast: forecasts,
		Runway: models.RunwayCalculation{
			MonthlyIncome:   math.Round(avgIncome*100) / 100,
			MonthlyExpenses: math.Round(avgExpenses*100) / 100,
			MonthlySavings:  math.Round(monthlySavings*100) / 100,
			CurrentBalance:  math.Round(accountsValue*100) / 100,
			RunwayMonths:    math.Round(runwayMonths*100) / 100,
		},
		UpcomingBills: bills,
	}, nil
}

type CashFlowResult struct {
	Forecast      []models.CashFlowForecast `json:"forecast"`
	Runway        models.RunwayCalculation  `json:"runway"`
	UpcomingBills []models.UpcomingBill     `json:"upcoming_bills"`
}

// GetAssetAllocation returns portfolio breakdown by asset type
func (s *AnalyticsService) GetAssetAllocation(userID uuid.UUID) ([]models.AssetAllocation, error) {
	holdings, err := s.holdingRepo.List(userID)
	if err != nil {
		return nil, err
	}

	var symbols []string
	for _, h := range holdings {
		symbols = append(symbols, h.Symbol)
	}
	prices, _ := s.holdingRepo.GetPrices(symbols)

	typeMap := make(map[string]*models.AssetAllocation)
	var total float64

	for _, h := range holdings {
		value := h.CostBasis
		if p, ok := prices[h.Symbol]; ok {
			value = p.Price * h.Quantity
		}
		at := string(h.AssetType)
		if _, ok := typeMap[at]; !ok {
			typeMap[at] = &models.AssetAllocation{AssetType: at}
		}
		typeMap[at].Value += value
		typeMap[at].Count++
		total += value
	}

	var result []models.AssetAllocation
	for _, a := range typeMap {
		if total > 0 {
			a.Percentage = (a.Value / total) * 100
		}
		result = append(result, *a)
	}
	if result == nil {
		result = []models.AssetAllocation{}
	}
	return result, nil
}

// calcMonthlyAverages fetches transactions, decrypts, and computes average monthly income/expense.
func (s *AnalyticsService) calcMonthlyAverages(userID uuid.UUID, months int) (income float64, expenses float64) {
	from := time.Now().AddDate(0, -months, 0)
	to := time.Now().AddDate(0, 0, 1) // tomorrow to include today

	txs, err := s.transactionRepo.ListByDateRange(userID, from, to)
	if err != nil {
		return 0, 0
	}

	monthlyIncome := make(map[string]float64)
	monthlyExpense := make(map[string]float64)

	for _, tx := range txs {
		if tx.Type != models.TransactionIncome && tx.Type != models.TransactionExpense {
			continue
		}
		plain, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(plain, 64)
		if err != nil {
			continue
		}
		period := formatPeriod(tx.Date)
		if tx.Type == models.TransactionIncome {
			monthlyIncome[period] += amt
		} else {
			monthlyExpense[period] += amt
		}
	}

	if len(monthlyIncome) > 0 {
		var total float64
		for _, v := range monthlyIncome {
			total += v
		}
		income = total / float64(len(monthlyIncome))
	}
	if len(monthlyExpense) > 0 {
		var total float64
		for _, v := range monthlyExpense {
			total += v
		}
		expenses = total / float64(len(monthlyExpense))
	}
	return
}

func parsePeriod(period string) (time.Time, time.Time) {
	now := time.Now()
	// Use tomorrow as the exclusive upper bound so that today's date is included
	// when date strings are used in queries (date < tomorrow).
	tomorrow := now.AddDate(0, 0, 1)
	switch period {
	case "week":
		from := now.AddDate(0, 0, -7)
		return from, tomorrow
	case "quarter":
		from := now.AddDate(0, -3, 0)
		return from, tomorrow
	case "year":
		from := now.AddDate(-1, 0, 0)
		return from, tomorrow
	case "ytd":
		from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return from, tomorrow
	default: // "month"
		from := now.AddDate(0, -1, 0)
		return from, tomorrow
	}
}

func monthlyMultiplier(freq string) float64 {
	switch freq {
	case "daily":
		return 30
	case "weekly":
		return 4.33
	case "biweekly":
		return 2.17
	case "monthly":
		return 1
	case "quarterly":
		return 1.0 / 3.0
	case "yearly":
		return 1.0 / 12.0
	default:
		return 0
	}
}

func formatPeriod(t time.Time) string {
	return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
}
