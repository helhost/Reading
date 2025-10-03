package main

import (
  "database/sql"
)

type Item struct {
  ID                int64   `json:"id"`
  Title             string  `json:"title"`
	Author            string  `json:"author"`
	NumChapters       int64   `json:"numChapters"`
	CompletedChapters int64   `json:"completedChapters"`
	Link              *string `json:"link,omitempty"`
}

// GetAllItems returns every row.
func GetAllItems(db *sql.DB) ([]Item, error) {
	rows, err := db.Query(`
		SELECT
			id,
			title,
			author,
			COALESCE(numChapters, 0),
			COALESCE(completedChapters, 0),
			link
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
    if err := rows.Scan(&it.ID, &it.Title, &it.Author, &it.NumChapters, &it.CompletedChapters, &it.Link); err != nil {
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
      title,
      author,
      COALESCE(numChapters, 0),
      COALESCE(completedChapters, 0),
      link
    FROM books
    WHERE id = ?;
  `, id)

  var it Item
  err := row.Scan(&it.ID, &it.Title, &it.Author, &it.NumChapters, &it.CompletedChapters, &it.Link)
  if err != nil {
    return Item{}, err
  }
  return it, nil
}


// AddBook inserts a new book and returns the created row.
func AddBook(db *sql.DB, title, author string, numChapters int64, link *string) (Item, error) {
	res, err := db.Exec(`
		INSERT INTO books (title, author, numChapters, completedChapters, link)
		VALUES (?, ?, ?, 0, ?);
	`, title, author, numChapters, link)
	if err != nil {
		return Item{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Item{}, err
	}

	// Reuse the existing reader to return the fully-populated record.
	return GetBookByID(db, id)
}


// UpdateCompletedChapters sets completedChapters for a book and returns the updated row.
func UpdateCompletedChapters(db *sql.DB, id int64, completed int64) (Item, error) {
	_, err := db.Exec(`
		UPDATE books
		SET completedChapters = ?
		WHERE id = ?;
	`, completed, id)
	if err != nil {
		return Item{}, err
	}
	return GetBookByID(db, id)
}
