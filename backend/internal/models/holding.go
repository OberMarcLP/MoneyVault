package models

import (
	"time"

	"github.com/google/uuid"
)

type AssetType string

const (
	AssetStock      AssetType = "stock"
	AssetETF        AssetType = "etf"
	AssetCrypto     AssetType = "crypto"
	AssetMutualFund AssetType = "mutual_fund"
	AssetDeFi       AssetType = "defi_position"
)

type Holding struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	AccountID    uuid.UUID `db:"account_id" json:"account_id"`
	AssetType    AssetType `db:"asset_type" json:"asset_type"`
	Symbol       string    `db:"symbol" json:"symbol"`
	Name         string    `db:"name" json:"name"`
	Quantity     float64   `db:"quantity" json:"quantity"`
	CostBasis    float64   `db:"cost_basis" json:"cost_basis"`
	Currency     string    `db:"currency" json:"currency"`
	AcquiredAt   time.Time `db:"acquired_at" json:"acquired_at"`
	Notes        string    `db:"notes" json:"notes"`
	Metadata     string    `db:"metadata" json:"metadata"`
	TokenAddress string    `db:"token_address" json:"token_address"`
	Network      string    `db:"network" json:"network"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"-"`
}

type HoldingWithPrice struct {
	Holding
	CurrentPrice  float64 `json:"current_price"`
	MarketValue   float64 `json:"market_value"`
	TotalReturn   float64 `json:"total_return"`
	ReturnPercent float64 `json:"return_percent"`
	DayChange     float64 `json:"day_change"`
	AssetName     string  `json:"asset_name"`
}

type CreateHoldingRequest struct {
	AccountID    string  `json:"account_id" binding:"required"`
	AssetType    string  `json:"asset_type" binding:"required,oneof=stock etf crypto mutual_fund defi_position"`
	Symbol       string  `json:"symbol" binding:"required"`
	Name         string  `json:"name"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
	CostBasis    float64 `json:"cost_basis" binding:"required,gte=0"`
	Currency     string  `json:"currency"`
	AcquiredAt   string  `json:"acquired_at" binding:"required"`
	Notes        string  `json:"notes"`
	TokenAddress string  `json:"token_address"`
	Network      string  `json:"network"`
	Metadata     string  `json:"metadata"`
}

type UpdateHoldingRequest struct {
	Quantity  *float64 `json:"quantity"`
	CostBasis *float64 `json:"cost_basis"`
	Notes     *string  `json:"notes"`
}

type PriceCache struct {
	ID            uuid.UUID `db:"id" json:"id"`
	Symbol        string    `db:"symbol" json:"symbol"`
	AssetType     string    `db:"asset_type" json:"asset_type"`
	Price         float64   `db:"price" json:"price"`
	Currency      string    `db:"currency" json:"currency"`
	ChangePercent float64   `db:"change_percent" json:"change_percent"`
	Name          string    `db:"name" json:"name"`
	FetchedAt     time.Time `db:"fetched_at" json:"fetched_at"`
}

type PriceHistory struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Symbol    string    `db:"symbol" json:"symbol"`
	AssetType string    `db:"asset_type" json:"asset_type"`
	Price     float64   `db:"price" json:"price"`
	Currency  string    `db:"currency" json:"currency"`
	Date      time.Time `db:"date" json:"date"`
	Source    string    `db:"source" json:"source"`
}

type TradeLot struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	HoldingID    uuid.UUID  `db:"holding_id" json:"holding_id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	Quantity     float64    `db:"quantity" json:"quantity"`
	CostPerUnit  float64    `db:"cost_per_unit" json:"cost_per_unit"`
	AcquiredAt   time.Time  `db:"acquired_at" json:"acquired_at"`
	SoldAt       *time.Time `db:"sold_at" json:"sold_at"`
	SoldPrice    *float64   `db:"sold_price" json:"sold_price"`
	SoldQuantity *float64   `db:"sold_quantity" json:"sold_quantity"`
	IsClosed     bool       `db:"is_closed" json:"is_closed"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type PortfolioSummary struct {
	TotalValue       float64 `json:"total_value"`
	TotalCost        float64 `json:"total_cost"`
	TotalReturn      float64 `json:"total_return"`
	TotalReturnPct   float64 `json:"total_return_pct"`
	DayChange        float64 `json:"day_change"`
	DayChangePct     float64 `json:"day_change_pct"`
	HoldingsCount    int     `json:"holdings_count"`
}

type CostBasisMethod string

const (
	CostBasisFIFO    CostBasisMethod = "fifo"
	CostBasisLIFO    CostBasisMethod = "lifo"
	CostBasisAverage CostBasisMethod = "average"
)

type SellHoldingRequest struct {
	Quantity float64         `json:"quantity" binding:"required,gt=0"`
	Price    float64         `json:"price" binding:"required,gt=0"`
	SoldAt   string          `json:"sold_at" binding:"required"`
	Method   CostBasisMethod `json:"method"`
}
