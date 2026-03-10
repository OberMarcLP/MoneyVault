package handlers

import (
	"net/http"

	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	repo *repositories.AuditRepository
}

func NewAuditHandler(repo *repositories.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

func (h *AuditHandler) List(c *gin.Context) {
	var filter models.AuditFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, total, err := h.repo.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs"})
		return
	}

	totalPages := total / filter.Limit
	if filter.Limit > 0 && total%filter.Limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":        logs,
		"total":       total,
		"page":        filter.Page,
		"total_pages": totalPages,
	})
}
