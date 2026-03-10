package handlers

import (
	"net/http"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BudgetHandler struct {
	service *services.BudgetService
}

func NewBudgetHandler(service *services.BudgetService) *BudgetHandler {
	return &BudgetHandler{service: service}
}

func (h *BudgetHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.service.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"budget": budget})
}

func (h *BudgetHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	budgets, err := h.service.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list budgets"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"budgets": budgets})
}

func (h *BudgetHandler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	budgetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	budget, err := h.service.GetByID(userID, budgetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"budget": budget})
}

func (h *BudgetHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	budgetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	var req models.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.service.Update(userID, budgetID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"budget": budget})
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	budgetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	if err := h.service.Delete(userID, budgetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete budget"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "budget deleted"})
}
