// @title Arvan SMS Gateway API
// @version 1.0
// @description API for sending SMS messages (Gateway Service).
// @BasePath /
package main

import (
	"arvan-sms-gateway/internal/api"
	"arvan-sms-gateway/internal/cache"
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/db"
	"arvan-sms-gateway/internal/logger"
	"arvan-sms-gateway/internal/metrics"
	"arvan-sms-gateway/internal/queue"
	"arvan-sms-gateway/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"strings"

	_ "arvan-sms-gateway/docs"
)

func main() {
	cfg := config.LoadEnv()

	logger.InitLogger()
	defer logger.Sync()
	metrics.InitMetrics()

	db.InitDB(cfg.DBUrl)
	service.InitService(cfg)
	cache.InitRedis(cfg.RedisAddr)

	r := gin.Default()

	// Only enable Swagger + CORS in Developer Mode
	if cfg.DeveloperMode == "true" {
		// CORS
		r.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(200)
				return
			}
			c.Next()
		})
		// Swagger
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api.RegisterRoutes(r, cfg)

	//jobs.StartRefundJob(db.DB, 60*time.Second, 10000) ##TODO is not scalable for just MVP

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	if err := queue.InitKafka(brokers); err != nil {
		logger.Error("Kafka producer init failed", zap.Error(err))
		panic(err)
	}
	defer queue.Close()

	logger.Info("Starting service",
		zap.String("service", cfg.ServiceName),
		zap.String("port", cfg.ServerPort),
	)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		logger.Error("Server failed", zap.Error(err))
		panic(err)
	}
}
