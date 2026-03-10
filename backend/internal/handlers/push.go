package handlers

import (
	"net/http"

	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PushHandler struct {
	service *services.PushService
}

func NewPushHandler(service *services.PushService) *PushHandler {
	return &PushHandler{service: service}
}

func (h *PushHandler) Subscribe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req models.PushSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Subscribe(userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscribed"})
}

func (h *PushHandler) Unsubscribe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Unsubscribe(userID, req.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed"})
}

func (h *PushHandler) GetVAPIDKey(c *gin.Context) {
	key := h.service.GetVAPIDPublicKey()
	c.JSON(http.StatusOK, gin.H{"public_key": key})
}
