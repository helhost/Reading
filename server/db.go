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
		CREATE TABLE IF NOT EXISTS books (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			numChapters INTEGER,
		  completedChapters INTEGER
		);
	`)
	return err
}
