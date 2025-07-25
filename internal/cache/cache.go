package cache

import (
	"context"
	"encoding/json"
	"time"

	"arvan-sms-gateway/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var ctx = context.Background()
var rdb *redis.Client

type UserData struct {
	IsVIP   bool  `json:"is_vip"`
	Balance int64 `json:"balance"`
}

func InitRedis(addr string) {
	rdb = redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect Redis", zap.Error(err))
		panic(err)
	}
	logger.Info("Connected to Redis")
}

// ------------------ Reservation ------------------

func SetReservation(userID string, amount int64, ttl time.Duration) {
	if rdb != nil {
		rdb.Set(ctx, "resv:"+userID, amount, ttl)
	}
}

func GetReservation(userID string) int64 {
	if rdb == nil {
		return 0
	}
	val, _ := rdb.Get(ctx, "resv:"+userID).Int64()
	return val
}

func DecrementReservation(userID string, cost int64) bool {
	if rdb == nil {
		return false
	}
	val, _ := rdb.Get(ctx, "resv:"+userID).Int64()
	if val < cost {
		return false
	}
	rdb.DecrBy(ctx, "resv:"+userID, cost)
	return true
}

func DeleteReservation(userID string) {
	if rdb != nil {
		rdb.Del(ctx, "resv:"+userID)
	}
}

// ------------------ User Cache ------------------

func SetUser(userID string, data *UserData, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "user:"+userID, jsonData, ttl).Err()
}

func GetUser(userID string) (*UserData, error) {
	if rdb == nil {
		return nil, nil
	}
	val, err := rdb.Get(ctx, "user:"+userID).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var data UserData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}
	return &data, nil
}
