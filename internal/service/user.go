package service

import (
	"arvan-sms-gateway/internal/cache"
	"arvan-sms-gateway/internal/db"
	"time"
)

type UserData struct {
	IsVIP   bool
	Balance int64
}

func GetUserData(userID string) (*UserData, error) {
	cached, err := cache.GetUser(userID)
	if err == nil && cached != nil {
		return &UserData{IsVIP: cached.IsVIP, Balance: cached.Balance}, nil
	}

	dbUser, err := db.GetUser(userID)
	if err != nil {
		return nil, err
	}

	data := &UserData{IsVIP: dbUser.IsVIP, Balance: dbUser.Balance}
	_ = cache.SetUser(userID, &cache.UserData{
		IsVIP:   dbUser.IsVIP,
		Balance: dbUser.Balance,
	}, 60*time.Second)

	return data, nil
}
