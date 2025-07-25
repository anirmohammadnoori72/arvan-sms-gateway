package queue

import (
	"arvan-sms-gateway/internal/logger"
	"arvan-sms-gateway/internal/metrics"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

var producer sarama.SyncProducer

func InitKafka(brokers []string) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Version = sarama.V3_6_0_0

	var err error
	producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		logger.Error("Kafka producer initialization failed", zap.Error(err))
		return err
	}
	logger.Info("Kafka producer initialized", zap.Strings("brokers", brokers))
	return nil
}

func SendMessage(topic string, key, value string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		logger.Error("Failed to send message to Kafka",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err),
		)
		metrics.KafkaErrors.Inc()
		return err
	}

	logger.Info("Message sent to Kafka",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)
	metrics.KafkaMessages.Inc()
	return nil
}

func Close() {
	if producer != nil {
		_ = producer.Close()
		logger.Info("Kafka producer closed")
	}
}
