package worker

import (
	"arvan-sms-gateway/internal/db"
	"arvan-sms-gateway/internal/logger"
	"arvan-sms-gateway/internal/metrics"
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
	config.Consumer.Offsets.AutoCommit.Enable = false

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

	logger.Info("Worker started",
		zap.String("topic", topic),
		zap.String("group", group),
		zap.Bool("isVIP", isVIP))

	for {
		if err := cg.Consume(ctx, []string{topic}, handler); err != nil {
			logger.Error("Kafka consume error", zap.Error(err))
			time.Sleep(2 * time.Second)
		}
		if ctx.Err() != nil {
			logger.Warn("Worker context canceled, shutting down")
			return
		}
	}
}

func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	logger.Info("Kafka consumer setup complete")
	return nil
}
func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Info("Kafka consumer cleanup complete")
	return nil
}

func (c *consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Info("Received Kafka message",
			zap.String("topic", claim.Topic()),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset))

		var req models.SMSRequest
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			logger.Error("Invalid Kafka message payload", zap.Error(err))
			sess.MarkMessage(msg, "")
			metrics.KafkaErrors.Inc()
			continue
		}

		if c.isVIP {
			handleVIP(req)
		} else {
			handleNormal(req)
		}

		metrics.KafkaMessages.Inc()
		metrics.QueueLength.Dec()

		sess.MarkMessage(msg, "")
		sess.Commit()
	}
	return nil
}
func handleVIP(req models.SMSRequest) {
	logger.Info("Processing VIP SMS",
		zap.String("message_id", req.MessageID),
		zap.String("user_id", req.UserID),
		zap.String("phone_number", req.PhoneNumber))

	sent := sendSMS(req.PhoneNumber, req.Message)
	if !sent {
		db.UpdateMessageStatus(req.MessageID, "failed")
		logger.Warn("VIP SMS failed", zap.String("message_id", req.MessageID))
		metrics.KafkaErrors.Inc()
		return
	}

	if err := db.DeductBalance(req.UserID, smsCost); err != nil {
		logger.Error("Failed to deduct balance", zap.Error(err))
		db.UpdateMessageStatus(req.MessageID, "error")
		return
	}

	db.UpdateMessageStatus(req.MessageID, "sent")
	logger.Info("VIP SMS sent successfully", zap.String("message_id", req.MessageID))
	metrics.TotalSMSRequests.Inc()
}

func handleNormal(req models.SMSRequest) {
	logger.Info("Processing Normal SMS",
		zap.String("message_id", req.MessageID),
		zap.String("user_id", req.UserID),
		zap.String("phone_number", req.PhoneNumber))

	// TODO: Reservation handling (MarkUsed or Rollback)
	sent := sendSMS(req.PhoneNumber, req.Message)

	if sent {
		db.UpdateMessageStatus(req.MessageID, "sent")
		logger.Info("Normal SMS sent successfully", zap.String("message_id", req.MessageID))
		metrics.TotalSMSRequests.Inc()
	} else {
		db.UpdateMessageStatus(req.MessageID, "failed")
		logger.Warn("Normal SMS failed", zap.String("message_id", req.MessageID))
		metrics.KafkaErrors.Inc()
	}
}

func sendSMS(phone, text string) bool {
	time.Sleep(10 * time.Millisecond)
	return true
}
