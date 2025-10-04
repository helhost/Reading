package store

import (
	"database/sql"
)

// User represents a user record in the database.
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"-"` // never JSON expose
	CreatedAt int64  `json:"created_at"`
}

// AddUser inserts a new user (already hashed password).
func AddUser(db *sql.DB, id, email, password string) error {
	_, err := db.Exec(`
		INSERT INTO users (id, email, password)
		VALUES (?, ?, ?);
	`, id, email, password)
	return err
}

// GetUserByEmail returns a user by their email.
func GetUserByEmail(db *sql.DB, email string) (User, error) {
	var u User
	err := db.QueryRow(`
		SELECT id, email, password, created_at
		FROM users
		WHERE email = ?;
	`, email).Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt)
	return u, err
}


// GetUserByID returns a user by their ID.
func GetUserByID(db *sql.DB, id string) (User, error) {
	var u User
	err := db.QueryRow(`
		SELECT id, email, password, created_at
		FROM users
		WHERE id = ?;
	`, id).Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt)
	return u, err
}
