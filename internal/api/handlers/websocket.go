package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "github.com/hendrowibowo/whrabbit/internal/api/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type WebSocketHandler struct {
	wsManager *ws.Manager
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		wsManager: ws.NewManager(),
	}
}

func NewWebSocketHandlerWithManager(manager *ws.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		wsManager: manager,
	}
}

func (h *WebSocketHandler) GetManager() *ws.Manager {
	return h.wsManager
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.wsManager.AddClient(conn)

	// Remove client when connection is closed
	defer h.wsManager.RemoveClient(conn)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
