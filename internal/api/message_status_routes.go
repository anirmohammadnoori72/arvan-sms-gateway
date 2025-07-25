package api

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary Get Message Status
// @Description Retrieve the delivery status of a previously submitted SMS by its Message ID.
// @Tags Messages
// @Produce  json
// @Param   message_id path string true "Message ID"
// @Success 200 {object} map[string]interface{} "Status retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid message ID"
// @Failure 404 {object} map[string]interface{} "Message not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /message-status/{message_id} [get]
func RegisterMessageStatusRoutes(r *gin.Engine, cfg *config.Config) {
	r.GET("/message-status/:message_id", func(c *gin.Context) {
		messageID := c.Param("message_id")
		status, err := db.GetMessageStatus(messageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch message status"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message_id": messageID,
			"status":     status,
		})
	})
}
