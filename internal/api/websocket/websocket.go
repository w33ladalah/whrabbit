package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Manager struct {
	clients    map[*websocket.Conn]bool
	clientsMux sync.RWMutex
	latestQR   string
	qrMux      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (m *Manager) AddClient(conn *websocket.Conn) {
	m.clientsMux.Lock()
	m.clients[conn] = true
	m.clientsMux.Unlock()

	// Send the latest QR code to the new client if available
	m.qrMux.RLock()
	if m.latestQR != "" {
		err := conn.WriteJSON(map[string]string{
			"type": "qr",
			"code": m.latestQR,
		})
		if err != nil {
			log.Printf("Error sending QR code to new client: %v", err)
		}
	}
	m.qrMux.RUnlock()
}

func (m *Manager) RemoveClient(conn *websocket.Conn) {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	delete(m.clients, conn)
	conn.Close()
}

func (m *Manager) BroadcastQR(qrCode string) {
	// Store the latest QR code
	m.qrMux.Lock()
	m.latestQR = qrCode
	m.qrMux.Unlock()

	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()

	for client := range m.clients {
		err := client.WriteJSON(map[string]string{
			"type": "qr",
			"code": qrCode,
		})
		if err != nil {
			log.Printf("Error sending QR code to client: %v", err)
			client.Close()
		}
	}
}
