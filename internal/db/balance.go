package db

import "database/sql"

func GetUserBalance(userID string) (int64, error) {
	var balance int64
	err := DB.QueryRow(`SELECT balance FROM users WHERE id = $1`, userID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func DeductBalance(userID string, amount int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance int64
	if err := tx.QueryRow(`SELECT balance FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&balance); err != nil {
		return err
	}
	if balance < amount {
		return sql.ErrNoRows
	}

	if _, err := tx.Exec(`UPDATE users SET balance=balance-$1 WHERE id=$2`, amount, userID); err != nil {
		return err
	}

	return tx.Commit()
}
