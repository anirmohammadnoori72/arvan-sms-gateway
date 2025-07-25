package db

import (
	"database/sql"

	"arvan-sms-gateway/internal/logger"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(url string) {
	var err error
	DB, err = sql.Open("postgres", url)
	if err != nil {
		logger.Error("Failed to connect to Postgres", zap.Error(err))
		panic(err)
	}
	if err = DB.Ping(); err != nil {
		logger.Error("Postgres not reachable", zap.Error(err))
		panic(err)
	}
	logger.Info("Connected to Postgres")
}

func IsVIPUser(userID string) (bool, error) {
	var isVIP bool
	err := DB.QueryRow(`SELECT is_vip FROM users WHERE id=$1`, userID).Scan(&isVIP)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return isVIP, nil
}
