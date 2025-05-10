package whatsapp

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/w33ladalah/whrabbit/internal/api/websocket"
	"github.com/w33ladalah/whrabbit/internal/config"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
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

	store.DeviceProps.Os = proto.String(config.GetAppName())
	store.DeviceProps.Version = &waProto.DeviceProps_AppVersion{
		Primary:   proto.Uint32(1),
		Secondary: proto.Uint32(0),
		Tertiary:  proto.Uint32(0),
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
				// Clear the device store to force new login
				waClient.Store.ID = nil
				// Trigger new QR code generation
				go func() {
					ctx := context.Background()
					qrChan, _ := waClient.GetQRChannel(ctx)
					err := waClient.Client.Connect()
					if err != nil {
						log.Printf("Error connecting after disconnect: %v", err)
						return
					}
					for evt := range qrChan {
						if evt.Event == "code" {
							waClient.wsManager.BroadcastQR(evt.Code)
							fmt.Println("New QR code after disconnect:", evt.Code)
						}
					}
				}()
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

// SendText sends a text message to a WhatsApp number
func (c *Client) SendText(to string, message string) error {
	recipient, err := ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient number: %v", err)
	}

	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	_, err = c.Client.SendMessage(context.Background(), recipient, msg)
	return err
}

// SendImage sends an image message to a WhatsApp number
func (c *Client) SendImage(to string, image io.Reader) error {
	recipient, err := ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient number: %v", err)
	}

	// Read image data
	imageData, err := io.ReadAll(image)
	if err != nil {
		return fmt.Errorf("error reading image: %v", err)
	}

	// Upload image to WhatsApp
	uploaded, err := c.Client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("error uploading image: %v", err)
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:        &uploaded.URL,
			Mimetype:   proto.String("image/jpeg"),
			Caption:    proto.String(""),
			FileSHA256: uploaded.FileSHA256,
			FileLength: &uploaded.FileLength,
			MediaKey:   uploaded.MediaKey,
			DirectPath: &uploaded.DirectPath,
			ViewOnce:   proto.Bool(false),
		},
	}

	_, err = c.Client.SendMessage(context.Background(), recipient, msg)
	return err
}
