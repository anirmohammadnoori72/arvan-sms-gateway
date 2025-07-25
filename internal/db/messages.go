package db

import (
	"arvan-sms-gateway/internal/models"
)

func InsertMessage(req models.SMSRequest, status string) error {
	_, err := DB.Exec(`
        INSERT INTO messages (message_id, user_id, phone_number, message, cost, status)
        VALUES ($1, $2, $3, $4, 1, $5)`,
		req.MessageID, req.UserID, req.PhoneNumber, req.Message, status)
	return err
}

func UpdateMessageStatus(messageID, status string) {
	DB.Exec(`UPDATE messages SET status = $1 WHERE message_id = $2`, status, messageID)
}

func GetMessageStatus(messageID string) (string, error) {
	var status string
	err := DB.QueryRow(`SELECT status FROM messages WHERE message_id = $1`, messageID).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}
