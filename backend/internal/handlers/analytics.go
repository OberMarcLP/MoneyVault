package handlers

import (
	"net/http"
	"strconv"

	"moneyvault/internal/middleware"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service *services.AnalyticsService
}

func NewAnalyticsHandler(service *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) NetWorthHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	days := 90
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	// Take a fresh snapshot first
	_ = h.service.TakeSnapshot(userID)

	data, err := h.service.GetNetWorthHistory(userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) SpendingBreakdown(c *gin.Context) {
	userID := middleware.GetUserID(c)
	period := c.DefaultQuery("period", "month")

	data, err := h.service.GetSpendingBreakdown(userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) SpendingTrends(c *gin.Context) {
	userID := middleware.GetUserID(c)
	months := 12
	if m := c.Query("months"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v > 0 {
			months = v
		}
	}

	data, err := h.service.GetSpendingTrends(userID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) TopExpenses(c *gin.Context) {
	userID := middleware.GetUserID(c)
	period := c.DefaultQuery("period", "month")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	data, err := h.service.GetTopExpenses(userID, period, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) BudgetVsActual(c *gin.Context) {
	userID := middleware.GetUserID(c)
	period := c.DefaultQuery("period", "month")

	data, err := h.service.GetBudgetVsActual(userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) CashFlowForecast(c *gin.Context) {
	userID := middleware.GetUserID(c)
	months := 6
	if m := c.Query("months"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v > 0 {
			months = v
		}
	}

	data, err := h.service.GetCashFlowForecast(userID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) AssetAllocation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	data, err := h.service.GetAssetAllocation(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
