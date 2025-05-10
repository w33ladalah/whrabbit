package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/w33ladalah/whrabbit/internal/api/handlers"
	"github.com/w33ladalah/whrabbit/internal/whatsapp"
)

// SetupRouter initializes the router and sets up all routes
func SetupRouter(waClient *whatsapp.Client, wsHandler *handlers.WebSocketHandler) *gin.Engine {
	router := gin.Default()

	// Create handlers
	messageHandler := handlers.NewMessageHandler(waClient)

	// Serve static files
	router.Static("/static", "static")
	router.StaticFile("/", "static/index.html")

	// WebSocket endpoint
	router.GET("/ws", wsHandler.HandleWebSocket)

	// API endpoints
	api := router.Group("/api")
	{
		// Message routes
		api.POST("/send/text", messageHandler.SendText)
		api.POST("/send/image", messageHandler.SendImage)
	}

	return router
}
