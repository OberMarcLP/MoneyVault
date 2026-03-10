package handlers

import (
	"log"
	"net/http"
	"strconv"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InvestmentHandler struct {
	service *services.InvestmentService
}

func NewInvestmentHandler(service *services.InvestmentService) *InvestmentHandler {
	return &InvestmentHandler{service: service}
}

func (h *InvestmentHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	holding, err := h.service.CreateHolding(userID, req)
	if err != nil {
		log.Printf("CreateHolding error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, holding)
}

func (h *InvestmentHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	holdings, err := h.service.ListHoldings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, holdings)
}

func (h *InvestmentHandler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	holding, err := h.service.GetHolding(userID, holdingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "holding not found"})
		return
	}
	c.JSON(http.StatusOK, holding)
}

func (h *InvestmentHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	var req models.UpdateHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	holding, err := h.service.UpdateHolding(userID, holdingID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, holding)
}

func (h *InvestmentHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	if err := h.service.DeleteHolding(userID, holdingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "holding deleted"})
}

func (h *InvestmentHandler) Sell(c *gin.Context) {
	userID := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	var req models.SellHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.SellHolding(userID, holdingID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sale recorded"})
}

func (h *InvestmentHandler) Summary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	summary, err := h.service.GetPortfolioSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *InvestmentHandler) RealizedGains(c *gin.Context) {
	userID := middleware.GetUserID(c)
	gains, err := h.service.GetRealizedGains(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gains)
}

func (h *InvestmentHandler) PriceHistory(c *gin.Context) {
	symbol := c.Param("symbol")
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 30
	}
	history, err := h.service.GetPriceHistory(symbol, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *InvestmentHandler) RefreshPrices(c *gin.Context) {
	if err := h.service.RefreshPrices(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "prices refreshed"})
}

// Dividend endpoints

func (h *InvestmentHandler) CreateDividend(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateDividendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d, err := h.service.CreateDividend(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *InvestmentHandler) ListDividends(c *gin.Context) {
	userID := middleware.GetUserID(c)
	holdingIDStr := c.Query("holding_id")

	if holdingIDStr != "" {
		holdingID, err := uuid.Parse(holdingIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid holding_id"})
			return
		}
		divs, err := h.service.ListDividendsByHolding(holdingID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"dividends": divs})
		return
	}

	divs, err := h.service.ListDividends(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dividends": divs})
}

func (h *InvestmentHandler) GetDividendSummary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	summary, err := h.service.GetDividendSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *InvestmentHandler) DeleteDividend(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dividend ID"})
		return
	}
	if err := h.service.DeleteDividend(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "dividend deleted"})
}
