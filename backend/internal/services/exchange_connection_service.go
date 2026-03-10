package services

import (
	"fmt"
	"log"
	"time"

	"moneyvault/internal/encryption"
	"moneyvault/internal/integrations"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

type ExchangeConnectionService struct {
	repo     *repositories.ExchangeConnectionRepository
	enc      *encryption.Service
	binance  *integrations.BinanceClient
	coinbase *integrations.CoinbaseClient
	kraken   *integrations.KrakenClient
}

func NewExchangeConnectionService(
	repo *repositories.ExchangeConnectionRepository,
	enc *encryption.Service,
	binance *integrations.BinanceClient,
	coinbase *integrations.CoinbaseClient,
	kraken *integrations.KrakenClient,
) *ExchangeConnectionService {
	return &ExchangeConnectionService{
		repo:     repo,
		enc:      enc,
		binance:  binance,
		coinbase: coinbase,
		kraken:   kraken,
	}
}

func (s *ExchangeConnectionService) Connect(userID uuid.UUID, req models.CreateExchangeConnectionRequest) (*models.ExchangeConnection, error) {
	// Encrypt API key and secret with user's DEK
	encKey, err := s.enc.EncryptField(userID, req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API key: %w", err)
	}
	encSecret, err := s.enc.EncryptField(userID, req.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API secret: %w", err)
	}

	conn := &models.ExchangeConnection{
		ID:        uuid.New(),
		UserID:    userID,
		Exchange:  models.ExchangeName(req.Exchange),
		APIKey:    encKey,
		APISecret: encSecret,
		Label:     req.Label,
		IsActive:  true,
	}

	if err := s.repo.Create(conn); err != nil {
		return nil, fmt.Errorf("failed to save exchange connection: %w", err)
	}

	// Clear sensitive fields before returning
	conn.APIKey = ""
	conn.APISecret = ""
	return conn, nil
}

func (s *ExchangeConnectionService) List(userID uuid.UUID) ([]models.ExchangeConnection, error) {
	conns, err := s.repo.List(userID)
	if err != nil {
		return nil, err
	}
	// Never return encrypted keys to the client
	for i := range conns {
		conns[i].APIKey = ""
		conns[i].APISecret = ""
	}
	return conns, nil
}

func (s *ExchangeConnectionService) Delete(userID, connID uuid.UUID) error {
	return s.repo.Delete(userID, connID)
}

func (s *ExchangeConnectionService) Sync(userID, connID uuid.UUID) (*models.ExchangeSyncResult, error) {
	conn, err := s.repo.GetByID(userID, connID)
	if err != nil {
		return nil, fmt.Errorf("exchange connection not found")
	}

	if !conn.IsActive {
		return nil, fmt.Errorf("exchange connection is disabled")
	}

	// Decrypt API credentials
	apiKey, err := s.enc.DecryptField(userID, conn.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}
	apiSecret, err := s.enc.DecryptField(userID, conn.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API secret: %w", err)
	}

	// Fetch balances from the exchange
	var balances []integrations.ExchangeBalanceResult
	switch conn.Exchange {
	case models.ExchangeBinance:
		balances, err = s.binance.GetBalances(apiKey, apiSecret)
	case models.ExchangeCoinbase:
		balances, err = s.coinbase.GetBalances(apiKey, apiSecret)
	case models.ExchangeKraken:
		balances, err = s.kraken.GetBalances(apiKey, apiSecret)
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", conn.Exchange)
	}

	if err != nil {
		log.Printf("Failed to sync %s for user %s: %v", conn.Exchange, userID, err)
		return nil, fmt.Errorf("failed to fetch balances from %s: %w", conn.Exchange, err)
	}

	// Update last synced timestamp
	_ = s.repo.UpdateLastSynced(connID)

	now := time.Now()
	result := &models.ExchangeSyncResult{
		ConnectionID: connID,
		Exchange:     conn.Exchange,
		SyncedAt:     now,
	}

	for _, b := range balances {
		result.Balances = append(result.Balances, models.ExchangeBalance{
			Symbol: b.Symbol,
			Free:   b.Free,
			Locked: b.Locked,
			Total:  b.Total,
		})
	}

	if result.Balances == nil {
		result.Balances = []models.ExchangeBalance{}
	}

	return result, nil
}

func (s *ExchangeConnectionService) ToggleActive(userID, connID uuid.UUID) error {
	conn, err := s.repo.GetByID(userID, connID)
	if err != nil {
		return fmt.Errorf("exchange connection not found")
	}
	return s.repo.SetActive(userID, connID, !conn.IsActive)
}
