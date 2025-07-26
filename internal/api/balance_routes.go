package api

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/db"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// @Summary Get User Balance
// @Description Retrieve the current wallet balance for a given user ID.
// @Tags Wallet
// @Produce  json
// @Param   user_id path string true "User ID"
// @Success 200 {object} map[string]interface{} "Balance retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID format"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /balance/{user_id} [get]
func RegisterBalanceRoutes(r *gin.Engine, cfg *config.Config) {
	r.GET("/balance/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")

		if _, err := uuid.Parse(userID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format (must be UUID)"})
			return
		}

		balance, err := db.GetUserBalance(userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"balance": balance,
		})
	})
}
