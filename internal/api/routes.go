package api

import (
	"arvan-sms-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	RegisterSMSRoutes(r, cfg)
	RegisterBalanceRoutes(r, cfg)
	RegisterMessageStatusRoutes(r, cfg)
}
