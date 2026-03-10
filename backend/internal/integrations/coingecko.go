package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CoinGeckoToken struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type CoinGeckoPrice struct {
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	CurrentPrice      float64 `json:"current_price"`
	PriceChange24h    float64 `json:"price_change_24h"`
	PriceChangePct24h float64 `json:"price_change_percentage_24h"`
	MarketCap         float64 `json:"market_cap"`
	Image             string  `json:"image"`
}

type CoinGeckoClient struct {
	httpClient *http.Client
	tokenList  []CoinGeckoToken
	tokenMap   map[string]string // symbol -> id
	mu         sync.RWMutex
	lastFetch  time.Time
}

func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		tokenMap:   make(map[string]string),
	}
}

func (c *CoinGeckoClient) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return c.httpClient.Do(req)
}

func (c *CoinGeckoClient) LoadTokenList() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastFetch) < 24*time.Hour && len(c.tokenList) > 0 {
		return nil
	}

	resp, err := c.doRequest("https://api.coingecko.com/api/v3/coins/list")
	if err != nil {
		return fmt.Errorf("coingecko token list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	var tokens []CoinGeckoToken
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return fmt.Errorf("failed to decode token list: %w", err)
	}

	c.tokenList = tokens
	c.tokenMap = make(map[string]string)
	for _, t := range tokens {
		key := strings.ToLower(t.Symbol)
		if _, exists := c.tokenMap[key]; !exists {
			c.tokenMap[key] = t.ID
		}
	}
	c.lastFetch = time.Now()
	return nil
}

func (c *CoinGeckoClient) SearchTokens(query string) []CoinGeckoToken {
	c.mu.RLock()
	defer c.mu.RUnlock()

	query = strings.ToLower(query)
	var results []CoinGeckoToken
	for _, t := range c.tokenList {
		if strings.Contains(strings.ToLower(t.Symbol), query) ||
			strings.Contains(strings.ToLower(t.Name), query) ||
			strings.Contains(strings.ToLower(t.ID), query) {
			results = append(results, t)
			if len(results) >= 20 {
				break
			}
		}
	}
	return results
}

func (c *CoinGeckoClient) ResolveID(symbol string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	symbol = strings.ToLower(symbol)
	well := map[string]string{
		"btc": "bitcoin", "eth": "ethereum", "sol": "solana",
		"ada": "cardano", "dot": "polkadot", "avax": "avalanche-2",
		"matic": "matic-network", "link": "chainlink", "uni": "uniswap",
		"aave": "aave", "doge": "dogecoin", "shib": "shiba-inu",
		"xrp": "ripple", "bnb": "binancecoin", "ltc": "litecoin",
		"atom": "cosmos", "near": "near", "apt": "aptos",
		"arb": "arbitrum", "op": "optimism", "usdt": "tether",
		"usdc": "usd-coin", "dai": "dai",
	}
	if id, ok := well[symbol]; ok {
		return id
	}
	if id, ok := c.tokenMap[symbol]; ok {
		return id
	}
	return symbol
}

func (c *CoinGeckoClient) GetPrices(ids []string) ([]CoinGeckoPrice, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&sparkline=false&price_change_percentage=24h",
		strings.Join(ids, ","),
	)

	resp, err := c.doRequest(url)
	if err != nil {
		return nil, fmt.Errorf("coingecko prices request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("coingecko rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	var prices []CoinGeckoPrice
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		return nil, fmt.Errorf("failed to decode prices: %w", err)
	}
	return prices, nil
}

func (c *CoinGeckoClient) GetPrice(id string) (*CoinGeckoPrice, error) {
	prices, err := c.GetPrices([]string{id})
	if err != nil {
		return nil, err
	}
	if len(prices) == 0 {
		return nil, fmt.Errorf("no price found for %s", id)
	}
	return &prices[0], nil
}
