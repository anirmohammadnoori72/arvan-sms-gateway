package api

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/models"
	"arvan-sms-gateway/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary Send SMS
// @Description Queue an SMS for delivery (via Kafka). Validates user, balance, phone number, and message size.
// @Tags SMS
// @Accept  json
// @Produce  json
// @Param   request body models.SMSRequest true "SMS Request"
// @Success 200 {object} map[string]interface{} "Message queued successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request, insufficient balance, invalid phone, or duplicate message ID"
// @Failure 429 {object} map[string]interface{} "Server busy, try again later"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /send-sms [post]
func RegisterSMSRoutes(r *gin.Engine, cfg *config.Config) {
	r.POST("/send-sms", func(c *gin.Context) {
		var req models.SMSRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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
