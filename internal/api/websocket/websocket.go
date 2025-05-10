package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Manager struct {
	clients     map[*websocket.Conn]bool
	clientsMux  sync.RWMutex
	latestQR    string
	qrMux       sync.RWMutex
	isConnected bool
	statusMux   sync.RWMutex
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

	// Check if already connected
	m.statusMux.RLock()
	if m.isConnected {
		err := conn.WriteJSON(map[string]string{
			"type":   "status",
			"status": "WhatsApp already connected!",
		})
		if err != nil {
			log.Printf("Error sending connection status to new client: %v", err)
		}
		m.statusMux.RUnlock()
		return
	}
	m.statusMux.RUnlock()

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

func (m *Manager) BroadcastConnectionStatus(status string) {
	m.statusMux.Lock()
	if status == "WhatsApp disconnected" {
		m.isConnected = false
		// Clear the latest QR code when disconnected
		m.qrMux.Lock()
		m.latestQR = ""
		m.qrMux.Unlock()
	} else {
		m.isConnected = status == "WhatsApp connected successfully!" || status == "WhatsApp already connected!"
	}
	m.statusMux.Unlock()

	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()

	for client := range m.clients {
		err := client.WriteJSON(map[string]string{
			"type":   "status",
			"status": status,
		})
		if err != nil {
			log.Printf("Error sending status to client: %v", err)
			client.Close()
		}
	}
}
