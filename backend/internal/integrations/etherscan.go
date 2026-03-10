package integrations

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

type EtherscanTx struct {
	Hash        string `json:"hash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	GasUsed     string `json:"gasUsed"`
	GasPrice    string `json:"gasPrice"`
	Timestamp   string `json:"timeStamp"`
	IsError     string `json:"isError"`
}

type etherscanResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []EtherscanTx `json:"result"`
}

// networkBaseURLs maps network names to their block explorer API base URLs.
// All use the Etherscan-compatible API format.
var networkBaseURLs = map[string]string{
	"ethereum": "https://api.etherscan.io/api",
	"polygon":  "https://api.polygonscan.com/api",
	"bsc":      "https://api.bscscan.com/api",
	"arbitrum": "https://api.arbiscan.io/api",
}

// NetworkNativeToken returns the native token symbol for a network.
func NetworkNativeToken(network string) string {
	switch network {
	case "polygon":
		return "MATIC"
	case "bsc":
		return "BNB"
	default:
		return "ETH"
	}
}

// NetworkExplorerURL returns the explorer URL for viewing a transaction.
func NetworkExplorerURL(network, txHash string) string {
	switch network {
	case "polygon":
		return "https://polygonscan.com/tx/" + txHash
	case "bsc":
		return "https://bscscan.com/tx/" + txHash
	case "arbitrum":
		return "https://arbiscan.io/tx/" + txHash
	default:
		return "https://etherscan.io/tx/" + txHash
	}
}

type EtherscanClient struct {
	httpClient *http.Client
}

func NewEtherscanClient() *EtherscanClient {
	return &EtherscanClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *EtherscanClient) GetTransactions(address string, startBlock int64) ([]EtherscanTx, error) {
	return c.GetTransactionsForNetwork("ethereum", address, startBlock)
}

func (c *EtherscanClient) GetTransactionsForNetwork(network, address string, startBlock int64) ([]EtherscanTx, error) {
	baseURL, ok := networkBaseURLs[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	url := fmt.Sprintf(
		"%s?module=account&action=txlist&address=%s&startblock=%d&endblock=99999999&sort=desc&page=1&offset=50",
		baseURL, address, startBlock,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("etherscan request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("etherscan returned status %d", resp.StatusCode)
	}

	var data etherscanResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode etherscan response: %w", err)
	}

	if data.Status != "1" && data.Message != "No transactions found" {
		return nil, fmt.Errorf("etherscan error: %s", data.Message)
	}

	return data.Result, nil
}

func WeiToEth(weiStr string) float64 {
	wei, err := strconv.ParseFloat(weiStr, 64)
	if err != nil {
		return 0
	}
	return wei / math.Pow(10, 18)
}

func GweiToEth(gweiStr string) float64 {
	gwei, err := strconv.ParseFloat(gweiStr, 64)
	if err != nil {
		return 0
	}
	return gwei / math.Pow(10, 9)
}
