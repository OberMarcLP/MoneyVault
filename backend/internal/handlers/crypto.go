package handlers

import (
	"net/http"
	"strconv"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CryptoHandler struct {
	service *services.CryptoService
}

func NewCryptoHandler(service *services.CryptoService) *CryptoHandler {
	return &CryptoHandler{service: service}
}

func (h *CryptoHandler) Summary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	summary, err := h.service.GetCryptoSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *CryptoHandler) SearchTokens(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' required"})
		return
	}
	tokens := h.service.SearchTokens(query)
	if tokens == nil {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *CryptoHandler) CreateWallet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wallet, err := h.service.CreateWallet(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wallet)
}

func (h *CryptoHandler) ListWallets(c *gin.Context) {
	userID := middleware.GetUserID(c)
	wallets, err := h.service.ListWallets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wallets)
}

func (h *CryptoHandler) DeleteWallet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}
	if err := h.service.DeleteWallet(userID, walletID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallet deleted"})
}

func (h *CryptoHandler) SyncWallet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}
	count, err := h.service.SyncWallet(userID, walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"synced": count})
}

func (h *CryptoHandler) WalletTransactions(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	txs, err := h.service.GetWalletTransactions(walletID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, txs)
}

func (h *CryptoHandler) AllWalletTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	txs, err := h.service.GetAllWalletTransactions(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, txs)
}

func (h *CryptoHandler) RefreshCryptoPrices(c *gin.Context) {
	if err := h.service.RefreshCryptoPrices(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "crypto prices refreshed"})
}
