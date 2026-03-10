package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"moneyvault/internal/integrations"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

type CryptoService struct {
	holdingRepo *repositories.HoldingRepository
	walletRepo  *repositories.WalletRepository
	coingecko   *integrations.CoinGeckoClient
	etherscan   *integrations.EtherscanClient
}

func NewCryptoService(
	holdingRepo *repositories.HoldingRepository,
	walletRepo *repositories.WalletRepository,
	coingecko *integrations.CoinGeckoClient,
	etherscan *integrations.EtherscanClient,
) *CryptoService {
	return &CryptoService{
		holdingRepo: holdingRepo,
		walletRepo:  walletRepo,
		coingecko:   coingecko,
		etherscan:   etherscan,
	}
}

func (s *CryptoService) SearchTokens(query string) []integrations.CoinGeckoToken {
	_ = s.coingecko.LoadTokenList()
	return s.coingecko.SearchTokens(query)
}

func (s *CryptoService) GetCryptoSummary(userID uuid.UUID) (*CryptoSummary, error) {
	holdings, err := s.holdingRepo.List(userID)
	if err != nil {
		return nil, err
	}

	var cryptoHoldings []models.Holding
	var defiHoldings []models.Holding
	for _, h := range holdings {
		if h.AssetType == models.AssetCrypto {
			cryptoHoldings = append(cryptoHoldings, h)
		} else if h.AssetType == models.AssetDeFi {
			defiHoldings = append(defiHoldings, h)
		}
	}

	var cryptoSymbols []string
	for _, h := range cryptoHoldings {
		cryptoSymbols = append(cryptoSymbols, h.Symbol)
	}

	prices, _ := s.holdingRepo.GetPrices(cryptoSymbols)

	summary := &CryptoSummary{}
	for _, h := range cryptoHoldings {
		if p, ok := prices[h.Symbol]; ok {
			summary.TotalValue += p.Price * h.Quantity
		}
		summary.TotalCost += h.CostBasis
		summary.TokenCount++
	}
	for _, h := range defiHoldings {
		if p, ok := prices[h.Symbol]; ok {
			summary.DeFiValue += p.Price * h.Quantity
		} else {
			summary.DeFiValue += h.CostBasis
		}
		summary.DeFiPositions++
	}
	summary.TotalReturn = summary.TotalValue - summary.TotalCost
	if summary.TotalCost > 0 {
		summary.TotalReturnPct = (summary.TotalReturn / summary.TotalCost) * 100
	}

	wallets, _ := s.walletRepo.List(userID)
	summary.WalletCount = len(wallets)

	gasFees, _ := s.walletRepo.GetTotalGasFees(userID)
	summary.TotalGasFees = gasFees

	return summary, nil
}

type CryptoSummary struct {
	TotalValue     float64 `json:"total_value"`
	TotalCost      float64 `json:"total_cost"`
	TotalReturn    float64 `json:"total_return"`
	TotalReturnPct float64 `json:"total_return_pct"`
	DeFiValue      float64 `json:"defi_value"`
	DeFiPositions  int     `json:"defi_positions"`
	TokenCount     int     `json:"token_count"`
	WalletCount    int     `json:"wallet_count"`
	TotalGasFees   float64 `json:"total_gas_fees"`
}

// Wallet management
func (s *CryptoService) CreateWallet(userID uuid.UUID, req models.CreateWalletRequest) (*models.Wallet, error) {
	network := req.Network
	if network == "" {
		network = "ethereum"
	}
	w := &models.Wallet{
		ID:      uuid.New(),
		UserID:  userID,
		Address: strings.ToLower(req.Address),
		Network: network,
		Label:   req.Label,
	}
	if err := s.walletRepo.Create(w); err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	return w, nil
}

func (s *CryptoService) ListWallets(userID uuid.UUID) ([]models.Wallet, error) {
	return s.walletRepo.List(userID)
}

func (s *CryptoService) DeleteWallet(userID, walletID uuid.UUID) error {
	return s.walletRepo.Delete(userID, walletID)
}

func (s *CryptoService) SyncWallet(userID, walletID uuid.UUID) (int, error) {
	wallet, err := s.walletRepo.GetByID(userID, walletID)
	if err != nil {
		return 0, err
	}

	txs, err := s.etherscan.GetTransactionsForNetwork(wallet.Network, wallet.Address, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	nativeToken := integrations.NetworkNativeToken(wallet.Network)

	count := 0
	for _, tx := range txs {
		if tx.IsError == "1" {
			continue
		}

		blockNum, _ := strconv.ParseInt(tx.BlockNumber, 10, 64)
		gasUsed, _ := strconv.ParseInt(tx.GasUsed, 10, 64)
		ts, _ := strconv.ParseInt(tx.Timestamp, 10, 64)

		gasPriceWei := integrations.WeiToEth(tx.GasPrice)
		gasFee := float64(gasUsed) * gasPriceWei

		txType := "receive"
		if strings.EqualFold(tx.From, wallet.Address) {
			txType = "send"
		}

		wtx := &models.WalletTransaction{
			ID:          uuid.New(),
			WalletID:    walletID,
			UserID:      userID,
			TxHash:      tx.Hash,
			BlockNumber: blockNum,
			FromAddress: tx.From,
			ToAddress:   tx.To,
			Value:       tx.Value,
			TokenSymbol: nativeToken,
			GasUsed:     gasUsed,
			GasPrice:    tx.GasPrice,
			GasFeeEth:   gasFee,
			TxType:      txType,
			Timestamp:   time.Unix(ts, 0),
		}

		if err := s.walletRepo.CreateTransaction(wtx); err == nil {
			count++
		}
	}

	_ = s.walletRepo.UpdateLastSynced(walletID)
	return count, nil
}

func (s *CryptoService) GetWalletTransactions(walletID uuid.UUID, limit int) ([]models.WalletTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.walletRepo.ListTransactions(walletID, limit)
}

func (s *CryptoService) GetAllWalletTransactions(userID uuid.UUID, limit int) ([]models.WalletTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.walletRepo.ListAllUserTransactions(userID, limit)
}

// Price refresh for crypto holdings via CoinGecko
func (s *CryptoService) RefreshCryptoPrices() error {
	symbols, err := s.holdingRepo.GetDistinctCryptoSymbols()
	if err != nil || len(symbols) == 0 {
		return err
	}

	_ = s.coingecko.LoadTokenList()

	var ids []string
	symbolToID := make(map[string]string)
	for _, sym := range symbols {
		id := s.coingecko.ResolveID(sym)
		ids = append(ids, id)
		symbolToID[id] = sym
	}

	for i := 0; i < len(ids); i += 25 {
		end := i + 25
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		prices, err := s.coingecko.GetPrices(batch)
		if err != nil {
			log.Printf("Failed to fetch crypto prices for %v: %v", batch, err)
			time.Sleep(2 * time.Second)
			continue
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		for _, p := range prices {
			sym := strings.ToUpper(p.Symbol)
			if original, ok := symbolToID[p.ID]; ok {
				sym = strings.ToUpper(original)
			}
			_ = s.holdingRepo.UpsertPrice(&models.PriceCache{
				Symbol:        sym,
				AssetType:     "crypto",
				Price:         p.CurrentPrice,
				Currency:      "USD",
				ChangePercent: p.PriceChangePct24h,
				Name:          p.Name,
			})
			_ = s.holdingRepo.InsertPriceHistory(&models.PriceHistory{
				Symbol:    sym,
				AssetType: "crypto",
				Price:     p.CurrentPrice,
				Currency:  "USD",
				Date:      today,
				Source:    "coingecko",
			})
		}

		if end < len(ids) {
			time.Sleep(1500 * time.Millisecond)
		}
	}
	return nil
}
