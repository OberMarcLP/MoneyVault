package handlers

import (
	"net/http"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, models.ErrValidation(err.Error()))
		return
	}

	tx, err := h.service.Create(userID, req)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"transaction": tx})
}

func (h *TransactionHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var filter models.TransactionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		middleware.RespondError(c, models.ErrValidation(err.Error()))
		return
	}

	result, err := h.service.List(userID, filter)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondError(c, models.ErrValidation("invalid transaction ID"))
		return
	}

	tx, err := h.service.GetByID(id, userID)
	if err != nil {
		middleware.RespondError(c, models.ErrNotFoundMsg("transaction not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction": tx})
}

func (h *TransactionHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondError(c, models.ErrValidation("invalid transaction ID"))
		return
	}

	var req models.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, models.ErrValidation(err.Error()))
		return
	}

	tx, err := h.service.Update(id, userID, req)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction": tx})
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondError(c, models.ErrValidation("invalid transaction ID"))
		return
	}

	if err := h.service.Delete(id, userID); err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "transaction deleted"})
}
