package config

import (
	"go.uber.org/zap"
	"os"
	"strconv"

	"arvan-sms-gateway/internal/logger"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		v, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return v
		}
		logger.Warn("Invalid int64 for ENV, using default",
			zap.String("key", key),
			zap.Int64("default", fallback),
		)
	}
	return fallback
}

type Config struct {
	ServerPort          string
	ServiceName         string
	KafkaBrokers        string
	KafkaTopic          string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DeveloperMode       string
	ServerMetricsPort   string
	DBUrl               string
	RedisAddr           string
	WalletTopic         string
	SMSTopic            string
	BatchSize           int64
	ReservationTTL      int64 // seconds
	UseRedisReservation bool
	KafkaVIPTopic       string
}

func LoadEnv() *Config {
	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8081"),
		ServiceName:         getEnv("SERVICE_NAME", "arvan-sms-gateway"),
		KafkaBrokers:        getEnv("KAFKA_BROKERS", "127.0.0.1:9092"),
		KafkaTopic:          getEnv("KAFKA_TOPIC", "sms-messages"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          getEnv("DB_PASSWORD", "password"),
		DBName:              getEnv("DB_NAME", "sms_db"),
		DeveloperMode:       getEnv("DEVELOPER_MODE", "true"),
		ServerMetricsPort:   getEnv("SERVER_METRICS_PORT", "9090"),
		DBUrl:               getEnv("DB_URL", "postgres://sms_user:sms_pass@localhost:5432/sms_gateway?sslmode=disable"),
		RedisAddr:           getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		WalletTopic:         getEnv("KAFKA_WALLET_TOPIC", "wallet-requests"),
		SMSTopic:            getEnv("KAFKA_SMS_TOPIC", "sms-messages"),
		BatchSize:           getEnvInt64("WALLET_BATCH_SIZE", 100),
		ReservationTTL:      getEnvInt64("WALLET_RESERVATION_TTL", 30),
		UseRedisReservation: getEnv("USE_REDIS_RESERVATION", "false") == "true",
		KafkaVIPTopic:       getEnv("KAFKA_TOPIC", "sms-messages"),
	}
}
