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
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,                     -- UUID
			email      TEXT NOT NULL UNIQUE,
			password   TEXT NOT NULL,                        -- bcrypt
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);

		CREATE TABLE IF NOT EXISTS courses (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			year     INTEGER NOT NULL CHECK (year BETWEEN 1900 AND 2100),
			term     INTEGER NOT NULL CHECK (term IN (1,2,3,4)),
			code     TEXT NOT NULL,
			name     TEXT NOT NULL,
			-- uniqueness is per-user
			UNIQUE (user_id, year, term, code)
		);

		CREATE INDEX IF NOT EXISTS idx_courses_user ON courses(user_id);

		CREATE TABLE IF NOT EXISTS books (
			id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id            TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title              TEXT NOT NULL,
			author             TEXT NOT NULL,
			numChapters        INTEGER,
			link               TEXT,
			course_id          INTEGER,
			FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL
		);

		CREATE INDEX IF NOT EXISTS idx_books_user ON books(user_id);
		CREATE INDEX IF NOT EXISTS idx_books_user_course ON books(user_id, course_id);


		CREATE TABLE IF NOT EXISTS progress (
				book_id       INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
				chapter_index INTEGER NOT NULL CHECK (chapter_index >= 1),
				PRIMARY KEY (book_id, chapter_index)
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
