package handlers

import (
	"net/http"

	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
)

type PasswordResetHandler struct {
	service *services.PasswordResetService
}

func NewPasswordResetHandler(service *services.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{service: service}
}

func (h *PasswordResetHandler) RequestReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.RequestReset(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process request"})
		return
	}

	if token == "" {
		// User not found, but return success to prevent email enumeration
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account with that email exists, a reset token has been generated.",
		})
		return
	}

	// Self-hosted: return the token directly (no email server)
	c.JSON(http.StatusOK, gin.H{
		"message": "Reset token generated. Use it to set a new password.",
		"token":   token,
	})
}

func (h *PasswordResetHandler) ConfirmReset(c *gin.Context) {
	var req models.PasswordResetConfirm
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully. Please sign in with your new password."})
}
