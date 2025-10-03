package main

import (
	"database/sql"
)

type Item struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	NumChapters int64 `json:"numChapters"`
	CompletedChapters int64 `json:"completedChapters"`
}

// GetAllItems returns every row.
func GetAllItems(db *sql.DB) ([]Item, error) {
	rows, err := db.Query(`
		SELECT
			id,
			name,
			COALESCE(numChapters, 0),
			COALESCE(completedChapters, 0)
		FROM books
		ORDER BY id;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Name, &it.NumChapters, &it.CompletedChapters); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}


// GetBookByID returns a single book by ID.
func GetBookByID(db *sql.DB, id int64) (Item, error) {
	row := db.QueryRow(`
		SELECT
			id,
			name,
			COALESCE(numChapters, 0),
			COALESCE(completedChapters, 0)
		FROM books
		WHERE id = ?;
	`, id)

	var it Item
	err := row.Scan(&it.ID, &it.Name, &it.NumChapters, &it.CompletedChapters)
	if err != nil {
		return Item{}, err
	}
	return it, nil
}
