package main

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/db"
	"arvan-sms-gateway/internal/logger"
	"arvan-sms-gateway/internal/worker"
	"strings"
	"time"
)

func main() {
	cfg := config.LoadEnv()
	logger.InitLogger()
	defer logger.Sync()
	db.InitDB(cfg.DBUrl)

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	groupID := "normal-worker-" + time.Now().Format("150405")
	worker.StartWorker(brokers, cfg.KafkaTopicNormal, groupID, false)
}
