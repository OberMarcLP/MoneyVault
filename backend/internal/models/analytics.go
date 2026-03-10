package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NetWorthSnapshot struct {
	ID               uuid.UUID       `db:"id" json:"id"`
	UserID           uuid.UUID       `db:"user_id" json:"user_id"`
	Date             time.Time       `db:"date" json:"date"`
	TotalValue       float64         `db:"total_value" json:"total_value"`
	AccountsValue    float64         `db:"accounts_value" json:"accounts_value"`
	InvestmentsValue float64         `db:"investments_value" json:"investments_value"`
	CryptoValue      float64         `db:"crypto_value" json:"crypto_value"`
	Breakdown        json.RawMessage `db:"breakdown" json:"breakdown"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
}

type SpendingByCategory struct {
	CategoryID    string  `db:"category_id" json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryColor string  `json:"category_color"`
	CategoryIcon  string  `json:"category_icon"`
	Total         float64 `db:"total" json:"total"`
	Count         int     `db:"count" json:"count"`
}

type SpendingTrend struct {
	Period string  `json:"period"`
	Income float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net    float64 `json:"net"`
}

type TopExpense struct {
	Description string  `db:"description" json:"description"`
	Amount      float64 `db:"amount" json:"amount"`
	Date        string  `db:"date" json:"date"`
	Category    string  `json:"category"`
}

type BudgetVsActual struct {
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryColor string  `json:"category_color"`
	BudgetAmount  float64 `json:"budget_amount"`
	ActualAmount  float64 `json:"actual_amount"`
	Difference    float64 `json:"difference"`
	Percentage    float64 `json:"percentage"`
}

type BudgetHistory struct {
	Period   string           `json:"period"`
	Budgets  []BudgetVsActual `json:"budgets"`
	TotalBud float64          `json:"total_budget"`
	TotalAct float64          `json:"total_actual"`
}

type CashFlowForecast struct {
	Period          string  `json:"period"`
	ProjectedIncome float64 `json:"projected_income"`
	ProjectedExpense float64 `json:"projected_expense"`
	NetCashFlow     float64 `json:"net_cash_flow"`
	RunningBalance  float64 `json:"running_balance"`
}

type UpcomingBill struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	DueDate     string  `json:"due_date"`
	Frequency   string  `json:"frequency"`
	AccountName string  `json:"account_name"`
}

type AssetAllocation struct {
	AssetType  string  `json:"asset_type"`
	Value      float64 `json:"value"`
	Percentage float64 `json:"percentage"`
	Count      int     `json:"count"`
}

type RunwayCalculation struct {
	MonthlySavings    float64 `json:"monthly_savings"`
	MonthlyIncome     float64 `json:"monthly_income"`
	MonthlyExpenses   float64 `json:"monthly_expenses"`
	CurrentBalance    float64 `json:"current_balance"`
	RunwayMonths      float64 `json:"runway_months"`
}
