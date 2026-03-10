package handlers

import (
	"fmt"
	"net/http"
	"time"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExportHandler struct {
	service *services.ExportService
}

func NewExportHandler(service *services.ExportService) *ExportHandler {
	return &ExportHandler{service: service}
}

func (h *ExportHandler) ExportTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	format := c.DefaultQuery("format", "csv")

	filter := models.TransactionFilter{}
	if from := c.Query("from"); from != "" {
		filter.DateFrom = &from
	}
	if to := c.Query("to"); to != "" {
		filter.DateTo = &to
	}
	if accID := c.Query("account_id"); accID != "" {
		id, err := uuid.Parse(accID)
		if err == nil {
			filter.AccountID = &id
		}
	}
	if txType := c.Query("type"); txType != "" {
		t := models.TransactionType(txType)
		filter.Type = &t
	}

	filename := fmt.Sprintf("transactions_%s", time.Now().Format("2006-01-02"))

	switch format {
	case "json":
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.json"`, filename))
		c.Header("Content-Type", "application/json")
		if err := h.service.ExportTransactionsJSON(userID, filter, c.Writer); err != nil {
			middleware.RespondError(c, models.ErrValidation(err.Error()))
		}
	default:
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, filename))
		c.Header("Content-Type", "text/csv")
		if err := h.service.ExportTransactionsCSV(userID, filter, c.Writer); err != nil {
			middleware.RespondError(c, models.ErrValidation(err.Error()))
		}
	}
}

func (h *ExportHandler) ExportAccounts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	format := c.DefaultQuery("format", "csv")

	filename := fmt.Sprintf("accounts_%s", time.Now().Format("2006-01-02"))

	switch format {
	case "json":
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.json"`, filename))
		c.Header("Content-Type", "application/json")
		if err := h.service.ExportAccountsJSON(userID, c.Writer); err != nil {
			middleware.RespondError(c, models.ErrValidation(err.Error()))
		}
	default:
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, filename))
		c.Header("Content-Type", "text/csv")
		if err := h.service.ExportAccountsCSV(userID, c.Writer); err != nil {
			middleware.RespondError(c, models.ErrValidation(err.Error()))
		}
	}
}

func (h *ExportHandler) ExportAll(c *gin.Context) {
	userID := middleware.GetUserID(c)

	allData, err := h.service.ExportAllJSON(userID)
	if err != nil {
		middleware.RespondError(c, models.ErrValidation(err.Error()))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="moneyvault_export_%s.json"`, time.Now().Format("2006-01-02")))
	c.JSON(http.StatusOK, allData)
}
