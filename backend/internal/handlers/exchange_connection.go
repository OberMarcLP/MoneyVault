package handlers

import (
	"net/http"

	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExchangeConnectionHandler struct {
	service *services.ExchangeConnectionService
}

func NewExchangeConnectionHandler(service *services.ExchangeConnectionService) *ExchangeConnectionHandler {
	return &ExchangeConnectionHandler{service: service}
}

func (h *ExchangeConnectionHandler) Connect(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req models.CreateExchangeConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := h.service.Connect(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conn)
}

func (h *ExchangeConnectionHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	conns, err := h.service.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conns)
}

func (h *ExchangeConnectionHandler) Sync(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	result, err := h.service.Sync(userID, connID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ExchangeConnectionHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	if err := h.service.Delete(userID, connID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "exchange connection deleted"})
}

func (h *ExchangeConnectionHandler) Toggle(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	if err := h.service.ToggleActive(userID, connID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "exchange connection toggled"})
}
