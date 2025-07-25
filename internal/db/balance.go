package db

func GetUserBalance(userID string) (int64, error) {
	var balance int64
	err := DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, userID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
