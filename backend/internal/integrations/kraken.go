package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const krakenBaseURL = "https://api.kraken.com"

type KrakenClient struct {
	httpClient *http.Client
}

func NewKrakenClient() *KrakenClient {
	return &KrakenClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type krakenBalanceResponse struct {
	Error  []string           `json:"error"`
	Result map[string]string  `json:"result"`
}

// krakenSignature creates the API-Sign header for Kraken private API endpoints.
func krakenSignature(urlPath, nonce, postData, secret string) (string, error) {
	secretDecoded, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid kraken secret: %w", err)
	}

	sha := sha256.New()
	sha.Write([]byte(nonce + postData))
	shaSum := sha.Sum(nil)

	mac := hmac.New(sha512.New, secretDecoded)
	mac.Write([]byte(urlPath))
	mac.Write(shaSum)

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// krakenSymbolToStandard converts Kraken's asset naming to standard symbols.
// Kraken prefixes with X/Z for crypto/fiat (e.g., XXBT -> BTC, ZUSD -> USD).
func krakenSymbolToStandard(symbol string) string {
	// Known mappings
	mapping := map[string]string{
		"XXBT":  "BTC",
		"XETH":  "ETH",
		"XXRP":  "XRP",
		"XLTC":  "LTC",
		"XXLM":  "XLM",
		"XXDG":  "DOGE",
		"ZUSD":  "USD",
		"ZEUR":  "EUR",
		"ZGBP":  "GBP",
		"ZJPY":  "JPY",
		"ZCAD":  "CAD",
		"ZAUD":  "AUD",
	}
	if mapped, ok := mapping[symbol]; ok {
		return mapped
	}
	// Strip X/Z prefix if 4+ chars and starts with X or Z
	if len(symbol) >= 4 && (symbol[0] == 'X' || symbol[0] == 'Z') {
		return symbol[1:]
	}
	return symbol
}

// GetBalances fetches all non-zero balances from Kraken.
func (c *KrakenClient) GetBalances(apiKey, apiSecret string) ([]ExchangeBalanceResult, error) {
	urlPath := "/0/private/Balance"
	nonce := strconv.FormatInt(time.Now().UnixMilli(), 10)

	postData := url.Values{}
	postData.Set("nonce", nonce)
	encodedData := postData.Encode()

	sig, err := krakenSignature(urlPath, nonce, encodedData, apiSecret)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", krakenBaseURL+urlPath, strings.NewReader(encodedData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("API-Key", apiKey)
	req.Header.Set("API-Sign", sig)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kraken request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kraken returned status %d", resp.StatusCode)
	}

	var data krakenBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode kraken response: %w", err)
	}

	if len(data.Error) > 0 {
		return nil, fmt.Errorf("kraken error: %s", strings.Join(data.Error, ", "))
	}

	var results []ExchangeBalanceResult
	for asset, balanceStr := range data.Result {
		balance, _ := strconv.ParseFloat(balanceStr, 64)
		if balance > 0 {
			results = append(results, ExchangeBalanceResult{
				Symbol: krakenSymbolToStandard(asset),
				Free:   balance,
				Total:  balance,
			})
		}
	}

	return results, nil
}
