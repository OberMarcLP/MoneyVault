package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const binanceBaseURL = "https://api.binance.com"

type BinanceClient struct {
	httpClient *http.Client
}

func NewBinanceClient() *BinanceClient {
	return &BinanceClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type binanceBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type binanceAccountResponse struct {
	Balances []binanceBalance `json:"balances"`
}

// signRequest creates an HMAC-SHA256 signature for Binance API authentication.
func signRequest(queryString, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

// GetBalances fetches all non-zero spot balances from Binance.
func (c *BinanceClient) GetBalances(apiKey, apiSecret string) ([]ExchangeBalanceResult, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	queryString := "timestamp=" + timestamp
	signature := signRequest(queryString, apiSecret)

	url := fmt.Sprintf("%s/api/v3/account?%s&signature=%s", binanceBaseURL, queryString, signature)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("binance request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance returned status %d", resp.StatusCode)
	}

	var account binanceAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode binance response: %w", err)
	}

	var results []ExchangeBalanceResult
	for _, b := range account.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		total := free + locked
		if total > 0 {
			results = append(results, ExchangeBalanceResult{
				Symbol: b.Asset,
				Free:   free,
				Locked: locked,
				Total:  total,
			})
		}
	}

	return results, nil
}
