package admin

import (
	"database/sql"
)

func IsAdmin(db *sql.DB, userID string) (bool, error) {
	var x int
	err := db.QueryRow(`SELECT 1 FROM admins WHERE user_id = ? LIMIT 1`, userID).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func CountUsers(db *sql.DB) (int64, error) {
	var n int64
	err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}
