package main

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/db"
	"arvan-sms-gateway/internal/logger"
	"arvan-sms-gateway/internal/worker"
	"strings"
)

func main() {
	cfg := config.LoadEnv()
	logger.InitLogger()
	defer logger.Sync()
	db.InitDB(cfg.DBUrl)

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	worker.StartWorker(brokers, cfg.KafkaTopic, "normal-worker-group", false)
}
