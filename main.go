package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/w33ladalah/whrabbit/docs"
	"github.com/w33ladalah/whrabbit/internal/api/handlers"
	"github.com/w33ladalah/whrabbit/internal/api/middleware"
	"github.com/w33ladalah/whrabbit/internal/config"
	"github.com/w33ladalah/whrabbit/internal/whatsapp"
)

// @title           Whrabbit WhatsApp API
// @version         1.0
// @description     An unofficial WhatsApp API built with Go and whatsmeow.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading environment variables:", err)
	}

	// Initialize WhatsApp client
	client, err := whatsapp.NewClient("whatsmeow.db")
	if err != nil {
		log.Fatalf("Error creating WhatsApp client: %v", err)
	}

	// Create WebSocket handler
	wsHandler := handlers.NewWebSocketHandler()
	client.SetWebSocketManager(wsHandler.GetManager())

	// Create message handler
	msgHandler := handlers.NewMessageHandler(client)

	// Initialize router
	router := gin.Default()

	fmt.Println("Base URL:", config.GetBaseURL())

	// Swagger documentation
	docs.SwaggerInfo.Title = "Whrabbit WhatsApp API"
	docs.SwaggerInfo.Description = "An unofficial WhatsApp API built with Go and whatsmeow."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = config.GetBaseURL()
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Serve static files
	router.Static("/static", "./static")

	// WebSocket endpoint
	router.GET("/ws", wsHandler.HandleWebSocket)

	// API routes with authentication
	api := router.Group("/api/v1")
	api.Use(middleware.APIKeyAuth())
	{
		// Message routes
		api.POST("/messages/text", msgHandler.SendText)
		api.POST("/messages/image", msgHandler.SendImage)
	}

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve index.html at root
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + config.GetServerPort(),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s...", config.GetServerPort())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Connect to WhatsApp
	go func() {
		if err := client.Connect(context.Background()); err != nil {
			log.Printf("Error connecting to WhatsApp: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
