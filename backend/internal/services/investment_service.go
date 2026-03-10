package services

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"moneyvault/internal/integrations"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

type InvestmentService struct {
	holdingRepo  *repositories.HoldingRepository
	dividendRepo *repositories.DividendRepository
	yahoo        *integrations.YahooClient
}

func NewInvestmentService(holdingRepo *repositories.HoldingRepository, dividendRepo *repositories.DividendRepository, yahoo *integrations.YahooClient) *InvestmentService {
	return &InvestmentService{holdingRepo: holdingRepo, dividendRepo: dividendRepo, yahoo: yahoo}
}

func (s *InvestmentService) CreateHolding(userID uuid.UUID, req models.CreateHoldingRequest) (*models.Holding, error) {
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID")
	}

	acquiredAt, err := time.Parse("2006-01-02", req.AcquiredAt)
	if err != nil {
		return nil, fmt.Errorf("invalid acquired_at date")
	}

	h := &models.Holding{
		ID:         uuid.New(),
		UserID:     userID,
		AccountID:  accountID,
		AssetType:  models.AssetType(req.AssetType),
		Symbol:     strings.ToUpper(req.Symbol),
		Name:       req.Name,
		Quantity:   req.Quantity,
		CostBasis:  req.CostBasis,
		Currency:   req.Currency,
		AcquiredAt: acquiredAt,
		Notes:      req.Notes,
	}
	if h.Currency == "" {
		h.Currency = "USD"
	}

	if h.Name == "" {
		if quote, err := s.yahoo.GetQuote(h.Symbol); err == nil && quote.Name != "" {
			h.Name = quote.Name
			_ = s.holdingRepo.UpsertPrice(&models.PriceCache{
				Symbol: quote.Symbol, AssetType: string(h.AssetType),
				Price: quote.Price, Currency: quote.Currency,
				ChangePercent: quote.ChangePercent, Name: quote.Name,
			})
		}
	}

	if err := s.holdingRepo.Create(h); err != nil {
		return nil, fmt.Errorf("failed to create holding: %w", err)
	}

	lot := &models.TradeLot{
		ID:          uuid.New(),
		HoldingID:   h.ID,
		UserID:      userID,
		Quantity:    req.Quantity,
		CostPerUnit: req.CostBasis / req.Quantity,
		AcquiredAt:  acquiredAt,
	}
	_ = s.holdingRepo.CreateTradeLot(lot)

	return h, nil
}

func (s *InvestmentService) ListHoldings(userID uuid.UUID) ([]models.HoldingWithPrice, error) {
	holdings, err := s.holdingRepo.List(userID)
	if err != nil {
		return nil, err
	}

	var symbols []string
	for _, h := range holdings {
		symbols = append(symbols, h.Symbol)
	}

	prices, _ := s.holdingRepo.GetPrices(symbols)

	var result []models.HoldingWithPrice
	for _, h := range holdings {
		hwp := models.HoldingWithPrice{Holding: h, AssetName: h.Name}

		if p, ok := prices[h.Symbol]; ok {
			hwp.CurrentPrice = p.Price
			hwp.MarketValue = p.Price * h.Quantity
			hwp.TotalReturn = hwp.MarketValue - h.CostBasis
			if h.CostBasis > 0 {
				hwp.ReturnPercent = (hwp.TotalReturn / h.CostBasis) * 100
			}
			hwp.DayChange = p.ChangePercent
			if p.Name != "" {
				hwp.AssetName = p.Name
			}
		}

		result = append(result, hwp)
	}

	if result == nil {
		result = []models.HoldingWithPrice{}
	}
	return result, nil
}

func (s *InvestmentService) GetHolding(userID, holdingID uuid.UUID) (*models.HoldingWithPrice, error) {
	h, err := s.holdingRepo.GetByID(userID, holdingID)
	if err != nil {
		return nil, err
	}

	hwp := &models.HoldingWithPrice{Holding: *h, AssetName: h.Name}
	if p, err := s.holdingRepo.GetPrice(h.Symbol); err == nil {
		hwp.CurrentPrice = p.Price
		hwp.MarketValue = p.Price * h.Quantity
		hwp.TotalReturn = hwp.MarketValue - h.CostBasis
		if h.CostBasis > 0 {
			hwp.ReturnPercent = (hwp.TotalReturn / h.CostBasis) * 100
		}
		hwp.DayChange = p.ChangePercent
		if p.Name != "" {
			hwp.AssetName = p.Name
		}
	}
	return hwp, nil
}

func (s *InvestmentService) UpdateHolding(userID, holdingID uuid.UUID, req models.UpdateHoldingRequest) (*models.Holding, error) {
	h, err := s.holdingRepo.GetByID(userID, holdingID)
	if err != nil {
		return nil, err
	}
	if req.Quantity != nil {
		h.Quantity = *req.Quantity
	}
	if req.CostBasis != nil {
		h.CostBasis = *req.CostBasis
	}
	if req.Notes != nil {
		h.Notes = *req.Notes
	}
	if err := s.holdingRepo.Update(h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s *InvestmentService) DeleteHolding(userID, holdingID uuid.UUID) error {
	return s.holdingRepo.Delete(userID, holdingID)
}

func (s *InvestmentService) SellHolding(userID, holdingID uuid.UUID, req models.SellHoldingRequest) error {
	h, err := s.holdingRepo.GetByID(userID, holdingID)
	if err != nil {
		return err
	}
	if req.Quantity > h.Quantity {
		return fmt.Errorf("sell quantity exceeds holding quantity")
	}

	soldAt, err := time.Parse("2006-01-02", req.SoldAt)
	if err != nil {
		return fmt.Errorf("invalid sold_at date")
	}

	method := req.Method
	if method == "" {
		method = models.CostBasisFIFO
	}

	// Get open lots in the appropriate order
	var lots []models.TradeLot
	switch method {
	case models.CostBasisLIFO:
		lots, _ = s.holdingRepo.GetOpenLotsDesc(holdingID)
	default: // FIFO and Average both iterate lots
		lots, _ = s.holdingRepo.GetOpenLots(holdingID)
	}

	if method == models.CostBasisAverage {
		// Average cost: distribute the sale proportionally across all open lots
		s.sellAverage(lots, req.Quantity, req.Price, soldAt)
	} else {
		// FIFO/LIFO: close lots in order
		s.sellInOrder(lots, req.Quantity, req.Price, soldAt)
	}

	// Calculate cost basis reduction based on method
	var costReduction float64
	if method == models.CostBasisAverage {
		// Average cost: reduce proportionally
		if h.Quantity > 0 {
			avgCost := h.CostBasis / h.Quantity
			costReduction = avgCost * req.Quantity
		}
	} else {
		// FIFO/LIFO: sum actual cost of the lots that were sold
		costReduction = s.calcLotCostSold(lots, req.Quantity)
	}

	h.Quantity -= req.Quantity
	h.CostBasis -= costReduction
	if h.Quantity <= 0 {
		h.Quantity = 0
		h.CostBasis = 0
	}
	return s.holdingRepo.Update(h)
}

// sellInOrder closes lots sequentially (used by FIFO and LIFO)
func (s *InvestmentService) sellInOrder(lots []models.TradeLot, qty, price float64, soldAt time.Time) {
	remaining := qty
	for i := range lots {
		if remaining <= 0 {
			break
		}
		lot := &lots[i]
		sell := math.Min(remaining, lot.Quantity)
		lot.SoldAt = &soldAt
		lot.SoldPrice = &price
		lot.SoldQuantity = &sell
		if sell >= lot.Quantity {
			lot.IsClosed = true
		}
		_ = s.holdingRepo.UpdateTradeLot(lot)
		remaining -= sell
	}
}

// sellAverage distributes the sale proportionally across all open lots
func (s *InvestmentService) sellAverage(lots []models.TradeLot, qty, price float64, soldAt time.Time) {
	var totalOpen float64
	for _, lot := range lots {
		totalOpen += lot.Quantity
	}
	if totalOpen <= 0 {
		return
	}

	remaining := qty
	for i := range lots {
		lot := &lots[i]
		// Proportional share of this lot
		share := (lot.Quantity / totalOpen) * qty
		sell := math.Min(share, lot.Quantity)
		sell = math.Min(sell, remaining)
		if sell <= 0 {
			continue
		}
		lot.SoldAt = &soldAt
		lot.SoldPrice = &price
		lot.SoldQuantity = &sell
		if sell >= lot.Quantity {
			lot.IsClosed = true
		}
		_ = s.holdingRepo.UpdateTradeLot(lot)
		remaining -= sell
	}
}

// calcLotCostSold calculates the total cost basis of units sold from lots
func (s *InvestmentService) calcLotCostSold(lots []models.TradeLot, qty float64) float64 {
	remaining := qty
	var cost float64
	for _, lot := range lots {
		if remaining <= 0 {
			break
		}
		sold := lot.SoldQuantity
		if sold == nil {
			continue
		}
		sell := math.Min(*sold, remaining)
		cost += sell * lot.CostPerUnit
		remaining -= sell
	}
	return cost
}

func (s *InvestmentService) GetPortfolioSummary(userID uuid.UUID) (*models.PortfolioSummary, error) {
	holdings, err := s.ListHoldings(userID)
	if err != nil {
		return nil, err
	}

	summary := &models.PortfolioSummary{HoldingsCount: len(holdings)}
	for _, h := range holdings {
		summary.TotalValue += h.MarketValue
		summary.TotalCost += h.CostBasis
		summary.DayChange += h.MarketValue * (h.DayChange / 100)
	}
	summary.TotalReturn = summary.TotalValue - summary.TotalCost
	if summary.TotalCost > 0 {
		summary.TotalReturnPct = (summary.TotalReturn / summary.TotalCost) * 100
	}
	if summary.TotalValue > 0 {
		summary.DayChangePct = (summary.DayChange / summary.TotalValue) * 100
	}
	return summary, nil
}

func (s *InvestmentService) GetRealizedGains(userID uuid.UUID) ([]models.TradeLot, error) {
	return s.holdingRepo.GetClosedLots(userID)
}

func (s *InvestmentService) GetPriceHistory(symbol string, days int) ([]models.PriceHistory, error) {
	return s.holdingRepo.GetPriceHistory(symbol, days)
}

func (s *InvestmentService) RefreshPrices() error {
	symbols, err := s.holdingRepo.GetDistinctSymbols()
	if err != nil || len(symbols) == 0 {
		return err
	}

	// Batch in groups of 10
	for i := 0; i < len(symbols); i += 10 {
		end := i + 10
		if end > len(symbols) {
			end = len(symbols)
		}
		batch := symbols[i:end]

		quotes, err := s.yahoo.GetQuotes(batch)
		if err != nil {
			log.Printf("Failed to fetch quotes for %v: %v", batch, err)
			continue
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		for _, q := range quotes {
			_ = s.holdingRepo.UpsertPrice(&models.PriceCache{
				Symbol:        q.Symbol,
				AssetType:     "stock",
				Price:         q.Price,
				Currency:      q.Currency,
				ChangePercent: q.ChangePercent,
				Name:          q.Name,
			})
			_ = s.holdingRepo.InsertPriceHistory(&models.PriceHistory{
				Symbol:    q.Symbol,
				AssetType: "stock",
				Price:     q.Price,
				Currency:  q.Currency,
				Date:      today,
				Source:    "yahoo",
			})
		}
	}
	return nil
}

// Dividend methods

func (s *InvestmentService) CreateDividend(userID uuid.UUID, req models.CreateDividendRequest) (*models.Dividend, error) {
	holdingID, err := uuid.Parse(req.HoldingID)
	if err != nil {
		return nil, fmt.Errorf("invalid holding ID")
	}

	// Verify holding belongs to user
	h, err := s.holdingRepo.GetByID(holdingID, userID)
	if err != nil || h == nil {
		return nil, fmt.Errorf("holding not found")
	}

	exDate, err := time.Parse("2006-01-02", req.ExDate)
	if err != nil {
		return nil, fmt.Errorf("invalid ex-date format, use YYYY-MM-DD")
	}

	currency := req.Currency
	if currency == "" {
		currency = h.Currency
	}

	d := &models.Dividend{
		ID:        uuid.New(),
		HoldingID: holdingID,
		UserID:    userID,
		Amount:    req.Amount,
		Currency:  currency,
		ExDate:    exDate,
		Notes:     req.Notes,
	}

	if req.PayDate != "" {
		pd, pErr := time.Parse("2006-01-02", req.PayDate)
		if pErr == nil {
			d.PayDate = &pd
		}
	}

	if err := s.dividendRepo.Create(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *InvestmentService) ListDividends(userID uuid.UUID) ([]models.Dividend, error) {
	return s.dividendRepo.ListByUser(userID)
}

func (s *InvestmentService) ListDividendsByHolding(holdingID, userID uuid.UUID) ([]models.Dividend, error) {
	return s.dividendRepo.ListByHolding(holdingID, userID)
}

func (s *InvestmentService) DeleteDividend(id, userID uuid.UUID) error {
	return s.dividendRepo.Delete(id, userID)
}

func (s *InvestmentService) GetDividendSummary(userID uuid.UUID) (*models.DividendSummary, error) {
	total, err := s.dividendRepo.GetTotalByUser(userID)
	if err != nil {
		return nil, err
	}

	ytd, err := s.dividendRepo.GetTotalYTD(userID, time.Now().Year())
	if err != nil {
		return nil, err
	}

	divs, err := s.dividendRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	summary := &models.DividendSummary{
		TotalDividends: total,
		DividendsYTD:   ytd,
		DividendCount:  len(divs),
	}

	if len(divs) > 0 {
		summary.LastDividendAt = divs[0].ExDate.Format("2006-01-02")
	}

	return summary, nil
}
