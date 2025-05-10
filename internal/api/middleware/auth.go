package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/w33ladalah/whrabbit/internal/config"
)

// APIKeyAuth middleware checks for valid API key
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedKey := config.GetAPIKey()

		if apiKey == "" || apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
