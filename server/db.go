package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// tiny app: keep connections low
	db.SetMaxOpenConns(1) // only 1 server should be active at a time
	db.SetMaxIdleConns(1)
	return db, nil
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS courses (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			year  INTEGER NOT NULL CHECK (year BETWEEN 1900 AND 2100),
			term  INTEGER NOT NULL CHECK (term IN (1,2,3,4)), -- e.g., 1=Winter,2=Spring,...
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
			link               TEXT,   -- optional, can be NULL
			course_id          INTEGER,
			FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL
		);
	`)
	return err
}
