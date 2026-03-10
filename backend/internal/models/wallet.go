package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	Address    string     `db:"address" json:"address"`
	Network    string     `db:"network" json:"network"`
	Label      string     `db:"label" json:"label"`
	IsActive   bool       `db:"is_active" json:"is_active"`
	LastSynced *time.Time `db:"last_synced" json:"last_synced"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type WalletTransaction struct {
	ID           uuid.UUID `db:"id" json:"id"`
	WalletID     uuid.UUID `db:"wallet_id" json:"wallet_id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	TxHash       string    `db:"tx_hash" json:"tx_hash"`
	BlockNumber  int64     `db:"block_number" json:"block_number"`
	FromAddress  string    `db:"from_address" json:"from_address"`
	ToAddress    string    `db:"to_address" json:"to_address"`
	Value        string    `db:"value" json:"value"`
	TokenSymbol  string    `db:"token_symbol" json:"token_symbol"`
	TokenAddress string    `db:"token_address" json:"token_address"`
	GasUsed      int64     `db:"gas_used" json:"gas_used"`
	GasPrice     string    `db:"gas_price" json:"gas_price"`
	GasFeeEth    float64   `db:"gas_fee_eth" json:"gas_fee_eth"`
	TxType       string    `db:"tx_type" json:"tx_type"`
	Timestamp    time.Time `db:"timestamp" json:"timestamp"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type CreateWalletRequest struct {
	Address string `json:"address" binding:"required"`
	Network string `json:"network"`
	Label   string `json:"label"`
}

type DeFiMetadata struct {
	Protocol     string  `json:"protocol"`
	PoolName     string  `json:"pool_name"`
	Token0       string  `json:"token0"`
	Token1       string  `json:"token1"`
	APY          float64 `json:"apy"`
	RewardsToken string  `json:"rewards_token"`
	PositionType string  `json:"position_type"`
}
