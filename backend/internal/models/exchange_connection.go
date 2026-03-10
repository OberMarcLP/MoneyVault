package models

import (
	"time"

	"github.com/google/uuid"
)

type ExchangeName string

const (
	ExchangeBinance  ExchangeName = "binance"
	ExchangeCoinbase ExchangeName = "coinbase"
	ExchangeKraken   ExchangeName = "kraken"
)

type ExchangeConnection struct {
	ID         uuid.UUID    `db:"id" json:"id"`
	UserID     uuid.UUID    `db:"user_id" json:"user_id"`
	Exchange   ExchangeName `db:"exchange" json:"exchange"`
	APIKey     string       `db:"api_key" json:"-"`
	APISecret  string       `db:"api_secret" json:"-"`
	Label      string       `db:"label" json:"label"`
	IsActive   bool         `db:"is_active" json:"is_active"`
	LastSynced *time.Time   `db:"last_synced" json:"last_synced"`
	CreatedAt  time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at" json:"updated_at"`
}

type CreateExchangeConnectionRequest struct {
	Exchange  string `json:"exchange" binding:"required,oneof=binance coinbase kraken"`
	APIKey    string `json:"api_key" binding:"required"`
	APISecret string `json:"api_secret" binding:"required"`
	Label     string `json:"label"`
}

type ExchangeBalance struct {
	Symbol   string  `json:"symbol"`
	Free     float64 `json:"free"`
	Locked   float64 `json:"locked"`
	Total    float64 `json:"total"`
	USDValue float64 `json:"usd_value,omitempty"`
}

type ExchangeSyncResult struct {
	ConnectionID uuid.UUID         `json:"connection_id"`
	Exchange     ExchangeName      `json:"exchange"`
	Balances     []ExchangeBalance `json:"balances"`
	SyncedAt     time.Time         `json:"synced_at"`
}
