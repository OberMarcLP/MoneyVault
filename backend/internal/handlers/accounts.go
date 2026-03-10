package handlers

import (
	"net/http"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccountHandler struct {
	service *services.AccountService
}

func NewAccountHandler(service *services.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

// Create godoc
// @Summary Create account
// @Description Creates a new financial account (checking, savings, credit, investment, crypto_wallet).
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateAccountRequest true "Account details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /accounts [post]
func (h *AccountHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, models.ErrValidation(err.Error()))
		return
	}

	account, err := h.service.Create(userID, req)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"account": account})
}

// List godoc
// @Summary List accounts
// @Description Returns all accounts for the authenticated user.
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /accounts [get]
func (h *AccountHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	accounts, err := h.service.List(userID)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

// GetByID godoc
// @Summary Get account by ID
// @Description Returns a single account by ID.
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Account UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /accounts/{id} [get]
func (h *AccountHandler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondError(c, models.ErrValidation("invalid account ID"))
		return
	}

	account, err := h.service.GetByID(id, userID)
	if err != nil {
		middleware.RespondError(c, models.ErrNotFoundMsg("account not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": account})
}

// Update godoc
// @Summary Update account
// @Description Updates an existing account.
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Account UUID"
// @Param request body models.UpdateAccountRequest true "Updated account details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /accounts/{id} [put]
func (h *AccountHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondError(c, models.ErrValidation("invalid account ID"))
		return
	}

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, models.ErrValidation(err.Error()))
		return
	}

	account, err := h.service.Update(id, userID, req)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": account})
}

func (h *AccountHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondError(c, models.ErrValidation("invalid account ID"))
		return
	}

	if err := h.service.Delete(id, userID); err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
