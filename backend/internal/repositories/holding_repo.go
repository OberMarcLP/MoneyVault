package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type HoldingRepository struct {
	db *sqlx.DB
}

func NewHoldingRepository(db *sqlx.DB) *HoldingRepository {
	return &HoldingRepository{db: db}
}

func (r *HoldingRepository) Create(h *models.Holding) error {
	query := `INSERT INTO holdings (id, user_id, account_id, asset_type, symbol, name, quantity, cost_basis, currency, acquired_at, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(query, h.ID, h.UserID, h.AccountID, h.AssetType, h.Symbol, h.Name,
		h.Quantity, h.CostBasis, h.Currency, h.AcquiredAt, h.Notes)
	return err
}

func (r *HoldingRepository) GetByID(userID, id uuid.UUID) (*models.Holding, error) {
	var h models.Holding
	err := r.db.Get(&h, `SELECT * FROM holdings WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return &h, err
}

func (r *HoldingRepository) List(userID uuid.UUID) ([]models.Holding, error) {
	var holdings []models.Holding
	err := r.db.Select(&holdings, `SELECT * FROM holdings WHERE user_id = $1 AND deleted_at IS NULL ORDER BY symbol ASC`, userID)
	if holdings == nil {
		holdings = []models.Holding{}
	}
	return holdings, err
}

func (r *HoldingRepository) Update(h *models.Holding) error {
	query := `UPDATE holdings SET quantity = $1, cost_basis = $2, notes = $3, name = $4, updated_at = NOW()
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL`
	_, err := r.db.Exec(query, h.Quantity, h.CostBasis, h.Notes, h.Name, h.ID, h.UserID)
	return err
}

func (r *HoldingRepository) Delete(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE holdings SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return err
}

func (r *HoldingRepository) GetDistinctSymbols() ([]string, error) {
	var symbols []string
	err := r.db.Select(&symbols, `SELECT DISTINCT symbol FROM holdings WHERE quantity > 0 AND deleted_at IS NULL`)
	if symbols == nil {
		symbols = []string{}
	}
	return symbols, err
}

func (r *HoldingRepository) GetDistinctCryptoSymbols() ([]string, error) {
	var symbols []string
	err := r.db.Select(&symbols, `SELECT DISTINCT symbol FROM holdings WHERE quantity > 0 AND asset_type IN ('crypto', 'defi_position') AND deleted_at IS NULL`)
	if symbols == nil {
		symbols = []string{}
	}
	return symbols, err
}

// Price Cache
func (r *HoldingRepository) UpsertPrice(p *models.PriceCache) error {
	query := `INSERT INTO price_cache (id, symbol, asset_type, price, currency, change_percent, name, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (symbol, asset_type) DO UPDATE SET price = $4, currency = $5, change_percent = $6, name = $7, fetched_at = NOW()`
	_, err := r.db.Exec(query, uuid.New(), p.Symbol, p.AssetType, p.Price, p.Currency, p.ChangePercent, p.Name)
	return err
}

func (r *HoldingRepository) GetPrice(symbol string) (*models.PriceCache, error) {
	var p models.PriceCache
	err := r.db.Get(&p, `SELECT * FROM price_cache WHERE symbol = $1`, symbol)
	return &p, err
}

func (r *HoldingRepository) GetPrices(symbols []string) (map[string]*models.PriceCache, error) {
	result := make(map[string]*models.PriceCache)
	if len(symbols) == 0 {
		return result, nil
	}
	query, args, err := sqlx.In(`SELECT * FROM price_cache WHERE symbol IN (?)`, symbols)
	if err != nil {
		return result, err
	}
	query = r.db.Rebind(query)
	var prices []models.PriceCache
	if err := r.db.Select(&prices, query, args...); err != nil {
		return result, err
	}
	for i := range prices {
		result[prices[i].Symbol] = &prices[i]
	}
	return result, nil
}

// Price History
func (r *HoldingRepository) InsertPriceHistory(p *models.PriceHistory) error {
	query := `INSERT INTO price_history (id, symbol, asset_type, price, currency, date, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (symbol, asset_type, date) DO NOTHING`
	_, err := r.db.Exec(query, uuid.New(), p.Symbol, p.AssetType, p.Price, p.Currency, p.Date, p.Source)
	return err
}

func (r *HoldingRepository) GetPriceHistory(symbol string, limit int) ([]models.PriceHistory, error) {
	var history []models.PriceHistory
	err := r.db.Select(&history, `SELECT * FROM price_history WHERE symbol = $1 ORDER BY date DESC LIMIT $2`, symbol, limit)
	if history == nil {
		history = []models.PriceHistory{}
	}
	return history, err
}

// Trade Lots
func (r *HoldingRepository) CreateTradeLot(lot *models.TradeLot) error {
	query := `INSERT INTO trade_lots (id, holding_id, user_id, quantity, cost_per_unit, acquired_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, lot.ID, lot.HoldingID, lot.UserID, lot.Quantity, lot.CostPerUnit, lot.AcquiredAt)
	return err
}

func (r *HoldingRepository) GetOpenLots(holdingID uuid.UUID) ([]models.TradeLot, error) {
	var lots []models.TradeLot
	err := r.db.Select(&lots, `SELECT * FROM trade_lots WHERE holding_id = $1 AND is_closed = false ORDER BY acquired_at ASC`, holdingID)
	if lots == nil {
		lots = []models.TradeLot{}
	}
	return lots, err
}

func (r *HoldingRepository) GetOpenLotsDesc(holdingID uuid.UUID) ([]models.TradeLot, error) {
	var lots []models.TradeLot
	err := r.db.Select(&lots, `SELECT * FROM trade_lots WHERE holding_id = $1 AND is_closed = false ORDER BY acquired_at DESC`, holdingID)
	if lots == nil {
		lots = []models.TradeLot{}
	}
	return lots, err
}

func (r *HoldingRepository) GetClosedLots(userID uuid.UUID) ([]models.TradeLot, error) {
	var lots []models.TradeLot
	err := r.db.Select(&lots, `SELECT * FROM trade_lots WHERE user_id = $1 AND is_closed = true ORDER BY sold_at DESC`, userID)
	if lots == nil {
		lots = []models.TradeLot{}
	}
	return lots, err
}

func (r *HoldingRepository) UpdateTradeLot(lot *models.TradeLot) error {
	query := `UPDATE trade_lots SET sold_at = $1, sold_price = $2, sold_quantity = $3, is_closed = $4 WHERE id = $5`
	_, err := r.db.Exec(query, lot.SoldAt, lot.SoldPrice, lot.SoldQuantity, lot.IsClosed, lot.ID)
	return err
}
