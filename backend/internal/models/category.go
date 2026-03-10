package models

import (
	"time"

	"github.com/google/uuid"
)

type CategoryType string

const (
	CategoryIncome  CategoryType = "income"
	CategoryExpense CategoryType = "expense"
)

type Category struct {
	ID        uuid.UUID    `db:"id" json:"id"`
	UserID    uuid.UUID    `db:"user_id" json:"user_id"`
	Name      string       `db:"name" json:"name"`
	Type      CategoryType `db:"type" json:"type"`
	Icon      string       `db:"icon" json:"icon"`
	Color     string       `db:"color" json:"color"`
	ParentID  *uuid.UUID   `db:"parent_id" json:"parent_id"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
}

type CreateCategoryRequest struct {
	Name     string       `json:"name" binding:"required"`
	Type     CategoryType `json:"type" binding:"required,oneof=income expense"`
	Icon     string       `json:"icon" binding:"required"`
	Color    string       `json:"color" binding:"required"`
	ParentID *uuid.UUID   `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Name     *string      `json:"name"`
	Type     *CategoryType `json:"type"`
	Icon     *string      `json:"icon"`
	Color    *string      `json:"color"`
	ParentID *uuid.UUID   `json:"parent_id"`
}

var DefaultCategories = []struct {
	Name  string
	Type  CategoryType
	Icon  string
	Color string
}{
	{"Salary", CategoryIncome, "briefcase", "#10B981"},
	{"Freelance", CategoryIncome, "laptop", "#059669"},
	{"Investments", CategoryIncome, "trending-up", "#0D9488"},
	{"Other Income", CategoryIncome, "plus-circle", "#14B8A6"},
	{"Housing", CategoryExpense, "home", "#6366F1"},
	{"Transportation", CategoryExpense, "car", "#8B5CF6"},
	{"Food & Dining", CategoryExpense, "utensils", "#EC4899"},
	{"Groceries", CategoryExpense, "shopping-cart", "#F43F5E"},
	{"Utilities", CategoryExpense, "zap", "#F59E0B"},
	{"Healthcare", CategoryExpense, "heart", "#EF4444"},
	{"Entertainment", CategoryExpense, "film", "#A855F7"},
	{"Shopping", CategoryExpense, "shopping-bag", "#D946EF"},
	{"Education", CategoryExpense, "book-open", "#3B82F6"},
	{"Insurance", CategoryExpense, "shield", "#0EA5E9"},
	{"Subscriptions", CategoryExpense, "repeat", "#06B6D4"},
	{"Travel", CategoryExpense, "plane", "#14B8A6"},
	{"Personal Care", CategoryExpense, "smile", "#F472B6"},
	{"Gifts & Donations", CategoryExpense, "gift", "#FB923C"},
	{"Taxes", CategoryExpense, "file-text", "#78716C"},
	{"Other Expenses", CategoryExpense, "more-horizontal", "#94A3B8"},
}
