package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// tiny app: keep connections low
	db.SetMaxOpenConns(1) // only 1 server should be active at a time
	db.SetMaxIdleConns(1)
	return db, nil
}


func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,                     -- UUID
			email      TEXT NOT NULL UNIQUE,
			password   TEXT NOT NULL,                        -- bcrypt or argon2 hash
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);

		CREATE TABLE IF NOT EXISTS courses (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			year  INTEGER NOT NULL CHECK (year BETWEEN 1900 AND 2100),
			term  INTEGER NOT NULL CHECK (term IN (1,2,3,4)),
			code  TEXT NOT NULL,
			name  TEXT NOT NULL,
			UNIQUE (year, term, code)
		);

		CREATE TABLE IF NOT EXISTS books (
			id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			title              TEXT NOT NULL,
			author             TEXT NOT NULL,
			numChapters        INTEGER,
			completedChapters  INTEGER,
			link               TEXT,
			course_id          INTEGER,
			FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id         TEXT PRIMARY KEY,                      -- opaque random token
			user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			expires_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_exp  ON sessions(expires_at);
	`)
	return err
}
