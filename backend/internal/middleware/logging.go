package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		userID := ""
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uuid.UUID); ok {
				userID = id.String()
			}
		}

		if userID != "" {
			log.Printf("[%s] %s %s %d %s user=%s", requestID[:8], method, path, status, duration, userID[:8])
		} else {
			log.Printf("[%s] %s %s %d %s", requestID[:8], method, path, status, duration)
		}
	}
}
