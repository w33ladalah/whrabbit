package whatsapp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hendrowibowo/whrabbit/internal/api/websocket"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Client wraps the WhatsApp client with additional functionality
type Client struct {
	*whatsmeow.Client
	wsManager *websocket.Manager
}

// NewClient creates a new WhatsApp client
func NewClient(dbPath string) (*Client, error) {
	container, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath), waLog.Stdout("Database", "DEBUG", true))
	if err != nil {
		return nil, fmt.Errorf("error creating database container: %v", err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, fmt.Errorf("error getting device store: %v", err)
	}

	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "DEBUG", true))
	waClient := &Client{
		Client:    client,
		wsManager: websocket.NewManager(),
	}

	// Add default event handler
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			log.Printf("Received message from %s: %s", v.Info.Sender, v.Message.GetConversation())
		case *events.Connected:
			log.Println("WhatsApp connected successfully")
			if waClient.wsManager != nil {
				waClient.wsManager.BroadcastConnectionStatus("WhatsApp connected successfully!")
			}
		case *events.Disconnected:
			log.Println("WhatsApp disconnected")
			if waClient.wsManager != nil {
				waClient.wsManager.BroadcastConnectionStatus("WhatsApp disconnected")
			}
		}
	})

	// Check if already logged in
	if deviceStore.ID != nil {
		waClient.wsManager.BroadcastConnectionStatus("WhatsApp already connected!")
	}

	return waClient, nil
}

// Connect connects to WhatsApp and handles QR code if needed
func (c *Client) Connect(ctx context.Context) error {
	if c.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := c.GetQRChannel(ctx)
		err := c.Client.Connect()
		if err != nil {
			return fmt.Errorf("error connecting to WhatsApp: %v", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Broadcast QR code to all connected WebSocket clients
				c.wsManager.BroadcastQR(evt.Code)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
				if evt.Event == "success" {
					c.wsManager.BroadcastConnectionStatus("WhatsApp connected successfully!")
				}
			}
		}
	} else {
		// Already logged in, just connect
		err := c.Client.Connect()
		if err != nil {
			return fmt.Errorf("error connecting to WhatsApp: %v", err)
		}
		c.wsManager.BroadcastConnectionStatus("WhatsApp already connected!")
	}
	return nil
}

// Disconnect disconnects from WhatsApp and triggers QR code generation
func (c *Client) Disconnect() {
	c.Client.Disconnect()
	c.wsManager.BroadcastConnectionStatus("WhatsApp disconnected")

	// Clear the device store to force new login
	c.Store.ID = nil

	// Trigger new QR code generation
	go func() {
		ctx := context.Background()
		qrChan, _ := c.GetQRChannel(ctx)
		err := c.Client.Connect()
		if err != nil {
			log.Printf("Error connecting after disconnect: %v", err)
			return
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				c.wsManager.BroadcastQR(evt.Code)
				fmt.Println("New QR code after disconnect:", evt.Code)
			}
		}
	}()
}

// SetWebSocketManager sets the WebSocket manager for the client
func (c *Client) SetWebSocketManager(manager *websocket.Manager) {
	c.wsManager = manager
}

// GetWebSocketManager returns the WebSocket manager for the client
func (c *Client) GetWebSocketManager() *websocket.Manager {
	return c.wsManager
}

// ParseJID parses a string into a JID
func ParseJID(arg string) (types.JID, error) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), nil
	} else {
		parts := strings.Split(arg, "@")
		if len(parts) != 2 {
			return types.JID{}, fmt.Errorf("invalid JID format")
		}
		return types.NewJID(parts[0], parts[1]), nil
	}
}
