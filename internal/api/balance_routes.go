package api

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary Get User Balance
// @Description Retrieve the current wallet balance for a given user ID.
// @Tags Wallet
// @Produce  json
// @Param   user_id path string true "User ID"
// @Success 200 {object} map[string]interface{} "Balance retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /balance/{user_id} [get]
func RegisterBalanceRoutes(r *gin.Engine, cfg *config.Config) {
	r.GET("/balance/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")
		balance, err := db.GetUserBalance(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch balance"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"balance": balance,
		})
	})
}
