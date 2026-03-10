package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type YahooQuote struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"regularMarketPrice"`
	Change        float64 `json:"regularMarketChange"`
	ChangePercent float64 `json:"regularMarketChangePercent"`
	Name          string  `json:"shortName"`
	Currency      string  `json:"currency"`
	MarketState   string  `json:"marketState"`
}

type stockPricesResponse struct {
	Ticker           string  `json:"Ticker"`
	Name             string  `json:"Name"`
	Price            float64 `json:"Price"`
	ChangeAmount     float64 `json:"ChangeAmount"`
	ChangePercentage float64 `json:"ChangePercentage"`
}

type YahooClient struct {
	httpClient *http.Client
}

func NewYahooClient() *YahooClient {
	return &YahooClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *YahooClient) fetchFromStockPrices(symbol, assetType string) (*YahooQuote, error) {
	endpoint := "stocks"
	if strings.EqualFold(assetType, "etf") {
		endpoint = "etfs"
	}

	url := fmt.Sprintf("https://stockprices.dev/api/%s/%s", endpoint, strings.ToUpper(symbol))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "MoneyVault/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stockprices.dev request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stockprices.dev returned status %d for %s", resp.StatusCode, symbol)
	}

	var data stockPricesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response for %s: %w", symbol, err)
	}

	if data.Price == 0 {
		return nil, fmt.Errorf("no price data for %s", symbol)
	}

	return &YahooQuote{
		Symbol:        strings.ToUpper(symbol),
		Price:         data.Price,
		Change:        data.ChangeAmount,
		ChangePercent: data.ChangePercentage,
		Name:          data.Name,
		Currency:      "USD",
		MarketState:   "REGULAR",
	}, nil
}

func (c *YahooClient) GetQuotes(symbols []string) ([]YahooQuote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	var quotes []YahooQuote
	var lastErr error

	for _, sym := range symbols {
		// Try as stock first, then ETF
		quote, err := c.fetchFromStockPrices(sym, "stock")
		if err != nil {
			quote, err = c.fetchFromStockPrices(sym, "etf")
		}
		if err != nil {
			lastErr = err
			continue
		}
		quotes = append(quotes, *quote)

		// Small delay to be respectful to the free API
		if len(symbols) > 1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	if len(quotes) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return quotes, nil
}

func (c *YahooClient) GetQuote(symbol string) (*YahooQuote, error) {
	quotes, err := c.GetQuotes([]string{symbol})
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote found for %s", symbol)
	}
	return &quotes[0], nil
}
