package api

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/models"
	"arvan-sms-gateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

// @Summary Send SMS
// @Description Queue an SMS for delivery (via Kafka). Validates user, balance, phone number, and message size.
// @Tags SMS
// @Accept  json
// @Produce  json
// @Param   request body models.SMSRequest true "SMS Request"
// @Success 200 {object} map[string]interface{} "Message queued successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request, invalid UUID, phone, or message size"
// @Failure 429 {object} map[string]interface{} "Server busy, try again later"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /send-sms [post]
func RegisterSMSRoutes(r *gin.Engine, cfg *config.Config) {
	r.POST("/send-sms", func(c *gin.Context) {
		var req models.SMSRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
			return
		}

		if _, err := uuid.Parse(req.UserID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format (must be UUID)"})
			return
		}
		if _, err := uuid.Parse(req.MessageID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id format (must be UUID)"})
			return
		}

		phone := strings.TrimSpace(req.PhoneNumber)
		if len(phone) < 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number"})
			return
		}

		if strings.TrimSpace(req.Message) == "" || len(req.Message) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message content"})
			return
		}

		result, err := service.ProcessSMSRequest(req, cfg)
		if err != nil {
			c.JSON(result.StatusCode, gin.H{"error": result.Message})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "pending",
			"message_id": req.MessageID,
		})
	})
}
