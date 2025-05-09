package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hendrowibowo/whrabbit/internal/whatsapp"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type MessageHandler struct {
	waClient *whatsapp.Client
}

func NewMessageHandler(waClient *whatsapp.Client) *MessageHandler {
	return &MessageHandler{
		waClient: waClient,
	}
}

func (h *MessageHandler) SendText(c *gin.Context) {
	var req struct {
		To      string `json:"to" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipient, err := whatsapp.ParseJID(req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient format"})
		return
	}

	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	_, err = h.waClient.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Message sent successfully"})
}

func (h *MessageHandler) SendImage(c *gin.Context) {
	// Get recipient number
	to := c.PostForm("to")
	if to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipient number is required"})
		return
	}

	// Get image file
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}
	defer file.Close()

	// Read image data
	imageData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image file"})
		return
	}

	// Get caption if provided
	caption := c.PostForm("caption")

	// Parse recipient JID
	recipient, err := whatsapp.ParseJID(to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient format"})
		return
	}

	// Upload image to WhatsApp servers
	uploadedImage, err := h.waClient.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload image: %v", err)})
		return
	}

	// Create message proto
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(caption),
			URL:           proto.String(uploadedImage.URL),
			DirectPath:    proto.String(uploadedImage.DirectPath),
			MediaKey:      uploadedImage.MediaKey,
			Mimetype:      proto.String(header.Header.Get("Content-Type")),
			FileLength:    proto.Uint64(uint64(len(imageData))),
			FileSHA256:    uploadedImage.FileSHA256,
			FileEncSHA256: uploadedImage.FileEncSHA256,
			JPEGThumbnail: nil, // Optional: You can add thumbnail support
		},
	}

	// Send message
	_, err = h.waClient.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send image: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Image sent successfully",
		"details": gin.H{
			"filename": header.Filename,
			"size":     header.Size,
			"caption":  caption,
		},
	})
}
