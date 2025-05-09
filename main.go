package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hendrowibowo/whrabbit/internal/api/handlers"
	"github.com/hendrowibowo/whrabbit/internal/api/routes"
	"github.com/hendrowibowo/whrabbit/internal/api/websocket"
	"github.com/hendrowibowo/whrabbit/internal/whatsapp"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	// Create WebSocket manager
	wsManager := websocket.NewManager()

	// Initialize WhatsApp client
	waClient, err := whatsapp.NewClient("whrabbit.db")
	if err != nil {
		log.Fatalf("Error initializing WhatsApp client: %v", err)
	}

	// Create WebSocket handler with the manager
	wsHandler := handlers.NewWebSocketHandlerWithManager(wsManager)

	// Set WebSocket manager for WhatsApp client
	waClient.SetWebSocketManager(wsManager)

	// Setup router
	router := routes.SetupRouter(waClient, wsHandler)

	// Start HTTP server
	log.Println("Starting server on http://localhost:8080")
	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Connect to WhatsApp in background
	go func() {
		if err := waClient.Connect(context.Background()); err != nil {
			log.Printf("Error connecting to WhatsApp: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down server...")
	waClient.Disconnect()
}
