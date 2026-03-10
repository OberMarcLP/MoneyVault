package middleware

import (
	"encoding/json"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Audit(auditRepo *repositories.AuditRepository, action string, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only log successful mutations (2xx status codes)
		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		var userID *uuid.UUID
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uuid.UUID); ok {
				userID = &id
			}
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		var resType *string
		if resourceType != "" {
			resType = &resourceType
		}

		// Try to get resource ID from URL param
		var resID *uuid.UUID
		if idStr := c.Param("id"); idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				resID = &id
			}
		}

		entry := &models.AuditLog{
			ID:           uuid.New(),
			UserID:       userID,
			Action:       action,
			ResourceType: resType,
			ResourceID:   resID,
			IPAddress:    &ip,
			UserAgent:    &ua,
		}

		_ = auditRepo.Create(entry)
	}
}

func AuditWithDetails(auditRepo *repositories.AuditRepository, action string, resourceType string, detailsFn func(c *gin.Context) json.RawMessage) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		var userID *uuid.UUID
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uuid.UUID); ok {
				userID = &id
			}
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		var resType *string
		if resourceType != "" {
			resType = &resourceType
		}

		var resID *uuid.UUID
		if idStr := c.Param("id"); idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				resID = &id
			}
		}

		var details *json.RawMessage
		if detailsFn != nil {
			d := detailsFn(c)
			if d != nil {
				details = &d
			}
		}

		entry := &models.AuditLog{
			ID:           uuid.New(),
			UserID:       userID,
			Action:       action,
			ResourceType: resType,
			ResourceID:   resID,
			IPAddress:    &ip,
			UserAgent:    &ua,
			Details:      details,
		}

		_ = auditRepo.Create(entry)
	}
}
