package middleware

import (
	"log"
	"moneyvault/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that catches panics and formats error responses consistently.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered: %v", r)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "an unexpected error occurred",
					},
				})
				c.Abort()
			}
		}()

		c.Next()

		// Format any AppError set via c.Error()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			if appErr, ok := err.(*models.AppError); ok {
				c.JSON(appErr.HTTPStatus, gin.H{
					"error": gin.H{
						"code":    appErr.Code,
						"message": appErr.Message,
					},
				})
				return
			}
		}
	}
}

// RespondError writes a standardized error JSON response from an error value.
// If the error is an *AppError, its code and status are used; otherwise defaults to 500.
func RespondError(c *gin.Context, err error) {
	if appErr, ok := err.(*models.AppError); ok {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "an unexpected error occurred",
		},
	})
}
