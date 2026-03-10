package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"moneyvault/internal/integrations"
)

type ExchangeHandler struct {
	client *integrations.ExchangeRateClient
}

func NewExchangeHandler(client *integrations.ExchangeRateClient) *ExchangeHandler {
	return &ExchangeHandler{client: client}
}

func (h *ExchangeHandler) GetRates(c *gin.Context) {
	base := c.DefaultQuery("base", "USD")

	rates, err := h.client.GetAllRates(base)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "exchange rates unavailable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"base":  base,
		"rates": rates,
	})
}

func (h *ExchangeHandler) Convert(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to currency codes required"})
		return
	}

	rate, err := h.client.GetRate(from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from": from,
		"to":   to,
		"rate": rate,
	})
}
