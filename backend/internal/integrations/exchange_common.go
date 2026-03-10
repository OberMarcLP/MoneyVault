package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// ExchangeBalanceResult represents a single asset balance from any exchange.
type ExchangeBalanceResult struct {
	Symbol string  `json:"symbol"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
	Total  float64 `json:"total"`
}

// hmacSHA256 computes HMAC-SHA256 and returns hex-encoded string.
func hmacSHA256(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
