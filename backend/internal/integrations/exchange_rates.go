package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ExchangeRateClient fetches exchange rates from the frankfurter.app API (ECB data, free, no key).
type ExchangeRateClient struct {
	httpClient *http.Client
	rates      map[string]float64 // currency -> rate relative to EUR
	base       string
	mu         sync.RWMutex
	lastFetch  time.Time
	cacheTTL   time.Duration
}

type frankfurterResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

func NewExchangeRateClient() *ExchangeRateClient {
	return &ExchangeRateClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		rates:      make(map[string]float64),
		base:       "EUR",
		cacheTTL:   1 * time.Hour,
	}
}

// FetchRates fetches latest exchange rates from frankfurter.app.
func (c *ExchangeRateClient) FetchRates() error {
	resp, err := c.httpClient.Get("https://api.frankfurter.app/latest")
	if err != nil {
		return fmt.Errorf("exchange rate fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("exchange rate API returned status %d", resp.StatusCode)
	}

	var data frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("exchange rate decode failed: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.rates = data.Rates
	c.rates[data.Base] = 1.0 // EUR = 1.0
	c.base = data.Base
	c.lastFetch = time.Now()

	return nil
}

// GetRate returns the exchange rate from one currency to another.
// Returns 1.0 if currencies are the same.
func (c *ExchangeRateClient) GetRate(from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.rates) == 0 {
		return 0, fmt.Errorf("exchange rates not loaded yet")
	}

	fromRate, okFrom := c.rates[from]
	toRate, okTo := c.rates[to]

	if !okFrom {
		return 0, fmt.Errorf("unsupported currency: %s", from)
	}
	if !okTo {
		return 0, fmt.Errorf("unsupported currency: %s", to)
	}

	// Convert: from -> EUR -> to
	// 1 FROM = (1/fromRate) EUR = (toRate/fromRate) TO
	return toRate / fromRate, nil
}

// Convert converts an amount from one currency to another.
func (c *ExchangeRateClient) Convert(amount float64, from, to string) (float64, error) {
	rate, err := c.GetRate(from, to)
	if err != nil {
		return 0, err
	}
	return amount * rate, nil
}

// GetAllRates returns all rates relative to the given base currency.
func (c *ExchangeRateClient) GetAllRates(base string) (map[string]float64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.rates) == 0 {
		return nil, fmt.Errorf("exchange rates not loaded yet")
	}

	baseRate, ok := c.rates[base]
	if !ok {
		return nil, fmt.Errorf("unsupported base currency: %s", base)
	}

	result := make(map[string]float64, len(c.rates))
	for currency, rate := range c.rates {
		result[currency] = rate / baseRate
	}

	return result, nil
}

// NeedsRefresh returns true if the cached rates are stale.
func (c *ExchangeRateClient) NeedsRefresh() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.lastFetch) > c.cacheTTL
}

// IsLoaded returns true if rates have been fetched at least once.
func (c *ExchangeRateClient) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.rates) > 0
}
