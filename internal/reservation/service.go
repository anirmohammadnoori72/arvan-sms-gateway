package reservation

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	db     *sql.DB
	rdb    *redis.Client
	bucket string
	ttl    time.Duration
}

func NewService(db *sql.DB, redisAddr string) *Service {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &Service{
		db:     db,
		rdb:    rdb,
		bucket: "wallet_tokens",
		ttl:    5 * time.Minute,
	}
}

func (s *Service) Reserve(userID string, tokens int64) (string, bool, error) {
	ctx := context.Background()
	key := s.bucket + ":" + userID

	val, err := s.rdb.DecrBy(ctx, key, tokens).Result()
	if err != nil && err != redis.Nil {
		return "", false, err
	}

	if val >= 0 {
		tx, err := s.db.Begin()
		if err != nil {
			return "", false, err
		}
		defer tx.Rollback()

		_, err = tx.Exec(`UPDATE users SET balance=balance-$1 WHERE id=$2`, tokens, userID)
		if err != nil {
			return "", false, err
		}

		resID := uuid.New().String()
		_, err = tx.Exec(`INSERT INTO reservations (id, user_id, amount, expires_at) VALUES ($1, $2, $3, NOW() + interval '5 minutes')`,
			resID, userID, tokens)
		if err != nil {
			return "", false, err
		}

		if err := tx.Commit(); err != nil {
			return "", false, err
		}

		return resID, true, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()

	var balance int64
	err = tx.QueryRow(`SELECT balance FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&balance)
	if err != nil {
		return "", false, err
	}
	if balance < tokens {
		return "", false, nil
	}

	_, err = tx.Exec(`UPDATE users SET balance=balance-$1 WHERE id=$2`, tokens, userID)
	if err != nil {
		return "", false, err
	}

	resID := uuid.New().String()
	_, err = tx.Exec(`INSERT INTO reservations (id, user_id, amount, expires_at) VALUES ($1, $2, $3, NOW() + interval '5 minutes')`,
		resID, userID, tokens)
	if err != nil {
		return "", false, err
	}

	if err := tx.Commit(); err != nil {
		return "", false, err
	}

	_ = s.rdb.Set(ctx, key, balance-tokens, s.ttl).Err()

	return resID, true, nil
}

func (s *Service) Rollback(userID string, tokens int64) error {
	ctx := context.Background()
	key := s.bucket + ":" + userID
	_ = s.rdb.IncrBy(ctx, key, tokens).Err()
	_, err := s.db.Exec(`UPDATE users SET balance=balance+$1 WHERE id=$2`, tokens, userID)
	return err
}

func (s *Service) MarkUsed(reservationID string) error {
	_, err := s.db.Exec(`UPDATE reservations SET used=true WHERE id=$1`, reservationID)
	return err
}
