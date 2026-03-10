package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const coinbaseBaseURL = "https://api.coinbase.com/v2"

type CoinbaseClient struct {
	httpClient *http.Client
}

func NewCoinbaseClient() *CoinbaseClient {
	return &CoinbaseClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type coinbaseAccount struct {
	ID       string `json:"id"`
	Currency struct {
		Code string `json:"code"`
	} `json:"currency"`
	Balance struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"balance"`
	NativeBalance struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"native_balance"`
}

type coinbaseAccountsResponse struct {
	Data       []coinbaseAccount `json:"data"`
	Pagination struct {
		NextURI string `json:"next_uri"`
	} `json:"pagination"`
}

// GetBalances fetches all non-zero account balances from Coinbase using API key auth.
func (c *CoinbaseClient) GetBalances(apiKey, apiSecret string) ([]ExchangeBalanceResult, error) {
	var allAccounts []coinbaseAccount

	path := "/v2/accounts?limit=100"
	for path != "" {
		url := "https://api.coinbase.com" + path

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		// Coinbase API v2 key auth (simplified — production would use CB-ACCESS-SIGN with HMAC)
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Set("CB-ACCESS-KEY", apiKey)
		req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
		req.Header.Set("CB-VERSION", "2024-01-01")
		req.Header.Set("Content-Type", "application/json")

		// HMAC-SHA256 signature: timestamp + method + path + body
		message := timestamp + "GET" + path
		sig := hmacSHA256(message, apiSecret)
		req.Header.Set("CB-ACCESS-SIGN", sig)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("coinbase request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("coinbase returned status %d", resp.StatusCode)
		}

		var data coinbaseAccountsResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("failed to decode coinbase response: %w", err)
		}

		allAccounts = append(allAccounts, data.Data...)

		if data.Pagination.NextURI != "" {
			path = data.Pagination.NextURI
		} else {
			path = ""
		}
	}

	var results []ExchangeBalanceResult
	for _, a := range allAccounts {
		amount, _ := strconv.ParseFloat(a.Balance.Amount, 64)
		if amount > 0 {
			results = append(results, ExchangeBalanceResult{
				Symbol: a.Currency.Code,
				Free:   amount,
				Total:  amount,
			})
		}
	}

	return results, nil
}
