package db

type User struct {
	ID      string
	Balance int64
	IsVIP   bool
}

func GetUser(userID string) (*User, error) {
	var u User
	err := DB.QueryRow(`SELECT id, balance, is_vip FROM users WHERE id=$1`, userID).
		Scan(&u.ID, &u.Balance, &u.IsVIP)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
