package reservation

import (
	"database/sql"
)

type PostgresReserver struct {
	db *sql.DB
}

func NewPostgresReserver(db *sql.DB) *PostgresReserver {
	return &PostgresReserver{db: db}
}

func (r *PostgresReserver) Reserve(userID string, tokens int) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var balance int
	err = tx.QueryRow(`SELECT balance FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&balance)
	if err != nil {
		return false, err
	}

	if balance < tokens {
		return false, nil
	}

	_, err = tx.Exec(`UPDATE users SET balance=balance-$1 WHERE id=$2`, tokens, userID)
	if err != nil {
		return false, err
	}

	return tx.Commit() == nil, nil
}

func (r *PostgresReserver) Commit(userID string, tokens int) error {
	// در این حالت نیازی نیست چون کم کردن موجودی در Reserve انجام شده
	return nil
}

func (r *PostgresReserver) Rollback(userID string, tokens int) error {
	_, err := r.db.Exec(`UPDATE users SET balance=balance+$1 WHERE id=$2`, tokens, userID)
	return err
}
