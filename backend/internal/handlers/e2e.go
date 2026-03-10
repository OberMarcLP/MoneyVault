package handlers

import (
	"net/http"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
)

type E2EHandler struct {
	service *services.E2EService
}

func NewE2EHandler(service *services.E2EService) *E2EHandler {
	return &E2EHandler{service: service}
}

// ExportData returns all user data with server-side decrypted fields for E2E migration.
func (h *E2EHandler) ExportData(c *gin.Context) {
	userID := middleware.GetUserID(c)
	data, err := h.service.ExportData(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// MigrateAndEnable atomically migrates data and enables E2E encryption.
func (h *E2EHandler) MigrateAndEnable(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Password        string                      `json:"password" binding:"required"`
		E2EEncryptedDEK string                      `json:"e2e_encrypted_dek" binding:"required"`
		E2EKEKSalt      string                      `json:"e2e_kek_salt" binding:"required"`
		Data            models.E2EMigrateDataRequest `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MigrateAndEnable(userID, req.Password, req.E2EEncryptedDEK, req.E2EKEKSalt, req.Data); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid password" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "E2E encryption enabled"})
}

// MigrateAndDisable re-encrypts data server-side and disables E2E encryption.
func (h *E2EHandler) MigrateAndDisable(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Password string                      `json:"password" binding:"required"`
		Data     models.E2EMigrateDataRequest `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MigrateAndDisable(userID, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "E2E encryption disabled"})
}
