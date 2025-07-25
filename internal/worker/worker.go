package worker

import (
	"arvan-sms-gateway/internal/db"
	"arvan-sms-gateway/internal/logger"
	"arvan-sms-gateway/internal/models"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const smsCost = 1

type consumer struct {
	isVIP bool
}

func StartWorker(brokers []string, topic, group string, isVIP bool) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	cg, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		logger.Error("Failed to create consumer group", zap.Error(err))
		return
	}
	defer cg.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		cancel()
	}()

	handler := &consumer{isVIP: isVIP}

	for {
		if err := cg.Consume(ctx, []string{topic}, handler); err != nil {
			logger.Error("Error from consumer", zap.Error(err))
			time.Sleep(2 * time.Second)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func (c *consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var req models.SMSRequest
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			logger.Error("Invalid message", zap.Error(err))
			sess.MarkMessage(msg, "")
			continue
		}

		if c.isVIP {
			handleVIP(req)
		} else {
			handleNormal(req)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func handleVIP(req models.SMSRequest) {
	sent := sendSMS(req.PhoneNumber, req.Message)
	if !sent {
		db.UpdateMessageStatus(req.MessageID, "failed")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}
	defer tx.Rollback()

	var balance int64
	err = tx.QueryRow(`SELECT balance FROM users WHERE id=$1 FOR UPDATE`, req.UserID).Scan(&balance)
	if err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}

	newBalance := balance - smsCost
	_, err = tx.Exec(`UPDATE users SET balance=$1 WHERE id=$2`, newBalance, req.UserID)
	if err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}

	if err := tx.Commit(); err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}

	db.UpdateMessageStatus(req.MessageID, "sent")
}

func handleNormal(req models.SMSRequest) {
	tx, err := db.DB.Begin()
	if err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}
	defer tx.Rollback()

	var balance int64
	err = tx.QueryRow(`SELECT balance FROM users WHERE id=$1 FOR UPDATE`, req.UserID).Scan(&balance)
	if err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}

	if balance < smsCost {
		db.UpdateMessageStatus(req.MessageID, "rejected")
		return
	}

	_, err = tx.Exec(`UPDATE users SET balance=balance-$1 WHERE id=$2`, smsCost, req.UserID)
	if err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}

	if err := tx.Commit(); err != nil {
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}

	sent := sendSMS(req.PhoneNumber, req.Message)
	if sent {
		db.UpdateMessageStatus(req.MessageID, "sent")
	} else {
		db.UpdateMessageStatus(req.MessageID, "failed")
	}
}

func sendSMS(phone, text string) bool {
	time.Sleep(10 * time.Millisecond)
	return true
}
