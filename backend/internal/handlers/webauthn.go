package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"

	"moneyvault/internal/middleware"
	"moneyvault/internal/services"
)

type WebAuthnHandler struct {
	service     *services.WebAuthnService
	authService *services.AuthService
}

func NewWebAuthnHandler(service *services.WebAuthnService, authService *services.AuthService) *WebAuthnHandler {
	return &WebAuthnHandler{service: service, authService: authService}
}

func (h *WebAuthnHandler) BeginRegistration(c *gin.Context) {
	userID := middleware.GetUserID(c)
	options, err := h.service.BeginRegistration(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"options": options})
}

func (h *WebAuthnHandler) FinishRegistration(c *gin.Context) {
	userID := middleware.GetUserID(c)

	response, err := protocol.ParseCredentialCreationResponseBody(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential response"})
		return
	}

	name := c.Query("name")
	if name == "" {
		name = "My Passkey"
	}

	if err := h.service.FinishRegistration(userID, name, response); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "passkey registered successfully"})
}

type beginLoginRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *WebAuthnHandler) BeginLogin(c *gin.Context) {
	var req beginLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	options, err := h.service.BeginLogin(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"options": options})
}

type finishLoginRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *WebAuthnHandler) FinishLogin(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	response, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential response"})
		return
	}

	accessToken, refreshToken, user, err := h.service.FinishLogin(email, response)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refresh_token", refreshToken,
		int(h.authService.RefreshTokenExpiry().Seconds()),
		"/api/v1/auth", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"user":         user,
	})
}

func (h *WebAuthnHandler) ListCredentials(c *gin.Context) {
	userID := middleware.GetUserID(c)
	creds, err := h.service.ListCredentials(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"credentials": creds})
}

func (h *WebAuthnHandler) DeleteCredential(c *gin.Context) {
	userID := middleware.GetUserID(c)
	credID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	if err := h.service.DeleteCredential(credID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "credential deleted"})
}
