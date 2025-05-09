package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hendrowibowo/whrabbit/internal/whatsapp"
)

type StatusHandler struct {
	waClient *whatsapp.Client
}

func NewStatusHandler(waClient *whatsapp.Client) *StatusHandler {
	return &StatusHandler{
		waClient: waClient,
	}
}

func (h *StatusHandler) GetStatus(c *gin.Context) {
	status := "disconnected"
	if h.waClient != nil && h.waClient.IsConnected() {
		status = "connected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"jid":    h.waClient.Store.ID.String(),
	})
}
