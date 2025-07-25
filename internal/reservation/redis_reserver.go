package reservation

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisReserver struct {
	rdb     *redis.Client
	bucket  string
	timeout time.Duration
}

func NewRedisReserver(addr string) *RedisReserver {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisReserver{
		rdb:     rdb,
		bucket:  "wallet_tokens",
		timeout: 2 * time.Second,
	}
}

func (r *RedisReserver) Reserve(userID string, tokens int) (bool, error) {
	ctx := context.Background()
	key := r.bucket + ":" + userID
	result := r.rdb.DecrBy(ctx, key, int64(tokens))
	val, err := result.Result()
	if err != nil {
		return false, err
	}
	if val < 0 {
		r.rdb.IncrBy(ctx, key, int64(tokens))
		return false, nil
	}
	return true, nil
}

func (r *RedisReserver) Commit(userID string, tokens int) error {
	// موجودی واقعی بعداً توسط Worker در Postgres کم می‌شود
	return nil
}

func (r *RedisReserver) Rollback(userID string, tokens int) error {
	ctx := context.Background()
	key := r.bucket + ":" + userID
	return r.rdb.IncrBy(ctx, key, int64(tokens)).Err()
}
