package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"moneyvault/internal/middleware"
	"moneyvault/internal/models"
	"moneyvault/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImportHandler struct {
	service *services.ImportService
}

func NewImportHandler(service *services.ImportService) *ImportHandler {
	return &ImportHandler{service: service}
}

func (h *ImportHandler) Preview(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	ext := strings.ToLower(header.Filename)
	var preview *models.CSVPreview

	switch {
	case strings.HasSuffix(ext, ".ofx") || strings.HasSuffix(ext, ".qfx"):
		preview, err = h.service.PreviewOFX(bytes.NewReader(data))
	case strings.HasSuffix(ext, ".qif"):
		preview, err = h.service.PreviewQIF(bytes.NewReader(data))
	default:
		preview, err = h.service.PreviewCSV(bytes.NewReader(data))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preview": preview})
}

func (h *ImportHandler) Import(c *gin.Context) {
	userID := middleware.GetUserID(c)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	accountIDStr := c.PostForm("account_id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}

	mapping := models.ColumnMapping{
		Date:        c.PostForm("map_date"),
		Amount:      c.PostForm("map_amount"),
		Description: c.PostForm("map_description"),
		Merchant:    c.PostForm("map_merchant"),
		Category:    c.PostForm("map_category"),
		SubCategory: c.PostForm("map_sub_category"),
		Type:        c.PostForm("map_type"),
		Status:      c.PostForm("map_status"),
		Currency:    c.PostForm("map_currency"),
	}
	if mapping.Date == "" || mapping.Amount == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date and amount column mappings required"})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	ext := strings.ToLower(header.Filename)
	var job *models.ImportJob

	switch {
	case strings.HasSuffix(ext, ".ofx") || strings.HasSuffix(ext, ".qfx"):
		job, err = h.service.ImportOFX(userID, accountID, bytes.NewReader(data), header.Filename)
	case strings.HasSuffix(ext, ".qif"):
		job, err = h.service.ImportQIF(userID, accountID, bytes.NewReader(data), header.Filename)
	default:
		postedOnly := strings.EqualFold(c.PostForm("posted_only"), "true") || c.PostForm("posted_only") == "1"
		job, err = h.service.ImportCSV(userID, accountID, bytes.NewReader(data), mapping, header.Filename, postedOnly)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"import": job})
}

func (h *ImportHandler) ListJobs(c *gin.Context) {
	userID := middleware.GetUserID(c)
	jobs, err := h.service.ListJobs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list imports"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"imports": jobs})
}
