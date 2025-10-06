package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"
)

type Session struct {
	ID        string
	UserID    string
	CreatedAt int64
	ExpiresAt int64
}

// CreateSession inserts a new session for userID with the given TTL.
func CreateSession(db *sql.DB, userID string, ttl time.Duration) (Session, error) {
	token, err := randomToken(32) // 256-bit
	if err != nil {
		return Session{}, err
	}
	now := time.Now().Unix()
	exp := time.Now().Add(ttl).Unix()

	_, err = db.Exec(`
		INSERT INTO sessions (id, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`, token, userID, now, exp)
	if err != nil {
		return Session{}, err
	}
	return Session{ID: token, UserID: userID, CreatedAt: now, ExpiresAt: exp}, nil
}

// GetSessionByID loads a session regardless of expiry.
// (Expiry is enforced by callers/middleware.)
func GetSessionByID(db *sql.DB, id string) (Session, error) {
	var s Session
	err := db.QueryRow(`
		SELECT id, user_id, created_at, expires_at
		FROM sessions
		WHERE id = ?;
	`, id).Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	return s, err
}

func DeleteSessionByID(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// Optional: opportunistic cleanup.
func DeleteExpiredSessions(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE expires_at <= strftime('%s','now')`)
	return err
}

func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// URL-safe, no padding
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
