package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "github.com/w33ladalah/whrabbit/internal/api/websocket"
	"github.com/w33ladalah/whrabbit/internal/whatsapp"
	"go.mau.fi/whatsmeow/types/events"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	manager *ws.Manager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		manager: ws.NewManager(),
	}
}

// NewWebSocketHandlerWithManager creates a new WebSocket handler with a custom manager
func NewWebSocketHandlerWithManager(manager *ws.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
	}
}

// GetManager returns the WebSocket manager
func (h *WebSocketHandler) GetManager() *ws.Manager {
	return h.manager
}

// HandleWebSocket handles WebSocket connections
// @Summary WebSocket connection for WhatsApp QR code and status updates
// @Description Establishes a WebSocket connection to receive WhatsApp QR codes and connection status updates
// @Tags websocket
// @Accept json
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Failure 500 {object} map[string]string "Error upgrading connection"
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	h.manager.AddClient(conn)

	// Handle client disconnection
	defer func() {
		h.manager.RemoveClient(conn)
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// MessageHandler handles WhatsApp messages
type MessageHandler struct {
	client *whatsapp.Client
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(client *whatsapp.Client) *MessageHandler {
	return &MessageHandler{
		client: client,
	}
}

// SendText sends a text message
// @Summary Send a text message
// @Description Sends a text message to a WhatsApp number
// @Tags messages
// @Accept json
// @Produce json
// @Param message body object true "Message details" SchemaExample({"to": "1234567890", "message": "Hello, World!"})
// @Success 200 {object} map[string]string "Message sent successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /messages/text [post]
func (h *MessageHandler) SendText(c *gin.Context) {
	var req struct {
		To      string `json:"to" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send the message using the WhatsApp client
	err := h.client.SendText(req.To, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Message sent successfully"})
}

// SendImage sends an image message
// @Summary Send an image message
// @Description Sends an image message to a WhatsApp number
// @Tags messages
// @Accept multipart/form-data
// @Produce json
// @Param to formData string true "Recipient's phone number"
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]string "Image sent successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /messages/image [post]
func (h *MessageHandler) SendImage(c *gin.Context) {
	to := c.PostForm("to")
	if to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipient number is required"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	// TODO: Implement image sending
	c.JSON(http.StatusOK, gin.H{"status": "Image sent"})
}

// NewEventHandler creates a new event handler function
func NewEventHandler(client *whatsapp.Client) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// Handle incoming messages
			log.Printf("Received message from %s: %s", v.Info.Sender, v.Message.GetConversation())
		case *events.Connected:
			// Handle successful connection
			if client.GetWebSocketManager() != nil {
				client.GetWebSocketManager().BroadcastConnectionStatus("WhatsApp connected successfully!")
			}
		case *events.Disconnected:
			// Handle disconnection
			if client.GetWebSocketManager() != nil {
				client.GetWebSocketManager().BroadcastConnectionStatus("WhatsApp disconnected")
			}
		}
	}
}
