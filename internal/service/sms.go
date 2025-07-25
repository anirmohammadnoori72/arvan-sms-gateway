package service

import (
	"arvan-sms-gateway/internal/config"
	"arvan-sms-gateway/internal/db"
	"arvan-sms-gateway/internal/logger"
	"arvan-sms-gateway/internal/models"
	"arvan-sms-gateway/internal/queue"
	"arvan-sms-gateway/internal/reservation"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type ServiceResult struct {
	StatusCode int
	Message    string
	MessageID  string
}

var reserverService *reservation.Service

func InitService(cfg *config.Config) {
	reserverService = reservation.NewService(db.DB, cfg.RedisAddr)
	logger.Info("Reservation service initialized with Redis + Postgres fallback")
}

func ProcessSMSRequest(req models.SMSRequest, cfg *config.Config) (*ServiceResult, error) {
	if req.MessageID == "" {
		return &ServiceResult{StatusCode: http.StatusBadRequest, Message: "message_id is required"}, nil
	}

	if err := db.InsertMessage(req, "queued"); err != nil {
		logger.Error("Failed to insert message", zap.Error(err))
		return &ServiceResult{StatusCode: http.StatusInternalServerError, Message: "db error"}, err
	}

	userData, err := GetUserData(req.UserID)
	if err != nil {
		logger.Error("Failed to fetch user", zap.Error(err))
		return &ServiceResult{StatusCode: http.StatusInternalServerError, Message: "user fetch error"}, err
	}

	topic := cfg.KafkaTopic
	if userData.IsVIP {
		topic = cfg.KafkaVIPTopic
	} else {
		_, ok, err := reserverService.Reserve(req.UserID, 1)
		if err != nil {
			logger.Error("Reservation error", zap.Error(err))
			return &ServiceResult{StatusCode: http.StatusInternalServerError, Message: "reservation error"}, err
		}
		if !ok {
			db.UpdateMessageStatus(req.MessageID, "rejected")
			return &ServiceResult{StatusCode: http.StatusBadRequest, Message: "insufficient balance"}, nil
		}
	}

	data, _ := json.Marshal(req)
	if err := queue.SendMessage(topic, req.UserID, string(data)); err != nil {
		logger.Error("Kafka enqueue error", zap.Error(err))
		if !userData.IsVIP {
			_ = reserverService.Rollback(req.UserID, 1)
		}
		db.UpdateMessageStatus(req.MessageID, "error")
		return &ServiceResult{StatusCode: http.StatusInternalServerError, Message: "kafka error"}, err
	}

	return &ServiceResult{
		StatusCode: http.StatusOK,
		Message:    "pending",
		MessageID:  req.MessageID,
	}, nil
}
