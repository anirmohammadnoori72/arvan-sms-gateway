package jobs

import (
	"database/sql"
	"time"

	"arvan-sms-gateway/internal/logger"
	"go.uber.org/zap"
)

func StartRefundJob(db *sql.DB, interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			refundExpired(db, batchSize)
		}
	}()
}

func refundExpired(db *sql.DB, batchSize int) {
	tx, err := db.Begin()
	if err != nil {
		logger.Error("refund tx start", zap.Error(err))
		return
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
        SELECT id, user_id, amount FROM reservations
        WHERE used=false AND expires_at < NOW()
        LIMIT $1 FOR UPDATE SKIP LOCKED`, batchSize)
	if err != nil {
		logger.Error("refund select", zap.Error(err))
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		var uid string
		var amt int64
		if err := rows.Scan(&id, &uid, &amt); err != nil {
			logger.Error("refund scan", zap.Error(err))
			return
		}
		_, err := tx.Exec(`UPDATE wallets SET balance=balance+$1 WHERE user_id=$2`, amt, uid)
		if err != nil {
			logger.Error("refund balance", zap.Error(err))
			return
		}
		ids = append(ids, id)
	}

	for _, id := range ids {
		_, _ = tx.Exec(`DELETE FROM reservations WHERE id=$1`, id)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("refund commit", zap.Error(err))
	} else {
		logger.Info("refund job done", zap.Int("count", len(ids)))
	}
}
