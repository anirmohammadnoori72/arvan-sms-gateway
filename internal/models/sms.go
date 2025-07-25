package models

import (
	"encoding/json"
)

type SMSRequest struct {
	UserID      string `json:"user_id"`
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	MessageID   string `json:"message_id"`
}

func (s *SMSRequest) ToJSON() string {
	data, _ := json.Marshal(s)
	return string(data)
}
