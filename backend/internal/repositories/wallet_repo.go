package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type WalletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(w *models.Wallet) error {
	query := `INSERT INTO wallets (id, user_id, address, network, label)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, w.ID, w.UserID, w.Address, w.Network, w.Label)
	return err
}

func (r *WalletRepository) GetByID(userID, id uuid.UUID) (*models.Wallet, error) {
	var w models.Wallet
	err := r.db.Get(&w, `SELECT * FROM wallets WHERE id = $1 AND user_id = $2`, id, userID)
	return &w, err
}

func (r *WalletRepository) List(userID uuid.UUID) ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := r.db.Select(&wallets, `SELECT * FROM wallets WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if wallets == nil {
		wallets = []models.Wallet{}
	}
	return wallets, err
}

func (r *WalletRepository) Delete(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM wallets WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *WalletRepository) UpdateLastSynced(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE wallets SET last_synced = NOW() WHERE id = $1`, id)
	return err
}

func (r *WalletRepository) CreateTransaction(tx *models.WalletTransaction) error {
	query := `INSERT INTO wallet_transactions (id, wallet_id, user_id, tx_hash, block_number,
		from_address, to_address, value, token_symbol, token_address,
		gas_used, gas_price, gas_fee_eth, tx_type, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (wallet_id, tx_hash) DO NOTHING`
	_, err := r.db.Exec(query, tx.ID, tx.WalletID, tx.UserID, tx.TxHash, tx.BlockNumber,
		tx.FromAddress, tx.ToAddress, tx.Value, tx.TokenSymbol, tx.TokenAddress,
		tx.GasUsed, tx.GasPrice, tx.GasFeeEth, tx.TxType, tx.Timestamp)
	return err
}

func (r *WalletRepository) ListTransactions(walletID uuid.UUID, limit int) ([]models.WalletTransaction, error) {
	var txs []models.WalletTransaction
	err := r.db.Select(&txs, `SELECT * FROM wallet_transactions WHERE wallet_id = $1 ORDER BY timestamp DESC LIMIT $2`, walletID, limit)
	if txs == nil {
		txs = []models.WalletTransaction{}
	}
	return txs, err
}

func (r *WalletRepository) ListAllUserTransactions(userID uuid.UUID, limit int) ([]models.WalletTransaction, error) {
	var txs []models.WalletTransaction
	err := r.db.Select(&txs, `SELECT * FROM wallet_transactions WHERE user_id = $1 ORDER BY timestamp DESC LIMIT $2`, userID, limit)
	if txs == nil {
		txs = []models.WalletTransaction{}
	}
	return txs, err
}

func (r *WalletRepository) GetTotalGasFees(userID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.Get(&total, `SELECT COALESCE(SUM(gas_fee_eth), 0) FROM wallet_transactions WHERE user_id = $1`, userID)
	return total, err
}
