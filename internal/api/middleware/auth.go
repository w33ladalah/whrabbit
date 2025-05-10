package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth middleware checks for valid API key
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedKey := os.Getenv("API_KEY")

		if apiKey == "" || apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
