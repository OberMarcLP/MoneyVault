package handlers

import (
	"net/http"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account. First user becomes admin.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "Registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _, err := h.service.Register(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// Login godoc
// @Summary Authenticate user
// @Description Login with email and password. Returns JWT access token and sets refresh token cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} map[string]string
// @Failure 428 {object} map[string]string "TOTP required"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, refreshToken, err := h.service.Login(req)
	if err != nil {
		if err.Error() == "totp_required" {
			c.JSON(http.StatusPreconditionRequired, gin.H{"error": "totp_required"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/api/v1/auth", "", false, true)
	c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Uses refresh token cookie to issue a new access token.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	accessToken, err := h.service.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

// Logout godoc
// @Summary Logout user
// @Description Revokes refresh token and clears DEK session.
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	refreshToken, _ := c.Cookie("refresh_token")
	h.service.Logout(userID, refreshToken)
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// Me godoc
// @Summary Get current user
// @Description Returns the authenticated user's profile.
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.service.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) UpdatePreferences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdatePreferences(userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "preferences updated"})
}

func (h *AuthHandler) SetupTOTP(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.service.SetupTOTP(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) VerifyTOTP(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.VerifyTOTP(userID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "TOTP enabled"})
}

func (h *AuthHandler) DisableTOTP(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.service.DisableTOTP(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "TOTP disabled"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.service.VerifyEmail(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "email verified"})
}

func (h *AuthHandler) EnableE2E(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.EnableE2ERequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.EnableE2E(userID, req.Password, req.E2EEncryptedDEK, req.E2EKEKSalt); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid password" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "E2E encryption enabled"})
}

func (h *AuthHandler) DisableE2E(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.DisableE2E(userID, req.Password); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid password" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "E2E encryption disabled"})
}
