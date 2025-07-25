package db

import (
	"time"

	"arvan-sms-gateway/internal/cache"
	"arvan-sms-gateway/internal/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func CreateReservation(userID string, batchSize int64, ttlSeconds int64) bool {
	tx, err := DB.Begin()
	if err != nil {
		logger.Error("DB transaction start failed", zap.Error(err))
		return false
	}
	defer tx.Rollback()

	res := tx.QueryRow(`
        UPDATE wallets
        SET balance = balance - $1
        WHERE user_id = $2 AND balance >= $1
        RETURNING balance`, batchSize, userID)

	var remaining int64
	if err := res.Scan(&remaining); err != nil {
		return false
	}

	id := uuid.New().String()
	expires := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	_, err = tx.Exec(`
        INSERT INTO reservations (id, user_id, amount, used, expires_at)
        VALUES ($1, $2, $3, 0, $4)`, id, userID, batchSize, expires)
	if err != nil {
		return false
	}

	if err = tx.Commit(); err != nil {
		logger.Error("DB commit failed", zap.Error(err))
		return false
	}

	cache.SetReservation(userID, batchSize, time.Duration(ttlSeconds)*time.Second)
	logger.Info("Reservation created",
		zap.String("user_id", userID),
		zap.Int64("batch", batchSize),
	)
	return true
}

func UseReservation(userID string, cost int64) bool {
	return cache.DecrementReservation(userID, cost)
}

func SyncReservationUsage(userID string, used int64) {
	DB.Exec(`
        UPDATE reservations
        SET used = GREATEST(used, $1)
        WHERE user_id = $2 AND expires_at > NOW()`, used, userID)
}

func ReconcileReservations() {
	now := time.Now()
	rows, _ := DB.Query(`
        SELECT id, user_id, amount, used
        FROM reservations
        WHERE expires_at <= $1`, now)
	defer rows.Close()

	for rows.Next() {
		var id, userID string
		var amount, used int64
		rows.Scan(&id, &userID, &amount, &used)
		if used < amount {
			remaining := amount - used
			DB.Exec(`UPDATE wallets SET balance = balance + $1 WHERE user_id = $2`, remaining, userID)
		}
		DB.Exec(`DELETE FROM reservations WHERE id = $1`, id)
		cache.DeleteReservation(userID)
	}
	logger.Info("Reconciliation job executed")
}
