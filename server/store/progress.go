package store

import (
	"database/sql"
)

// ListProgress returns the completed chapter indices for a book owned by user.
func ListProgress(db *sql.DB, userID string, bookID int64) ([]int64, error) {
	var numChapters int64
	if err := db.QueryRow(`
		SELECT COALESCE(numChapters, 0)
		  FROM books
		 WHERE id = ? AND user_id = ?;
	`, bookID, userID).Scan(&numChapters); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	rows, err := db.Query(`
		SELECT chapter_index
		  FROM progress
		 WHERE book_id = ?
		 ORDER BY chapter_index;
	`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []int64
	for rows.Next() {
		var idx int64
		if err := rows.Scan(&idx); err != nil {
			return nil, err
		}
		out = append(out, idx)
	}
	return out, rows.Err()
}

// AddProgress marks a single chapter as completed for a book owned by user.
// Idempotent: inserting an existing (book_id, chapter_index) is ignored.
func AddProgress(db *sql.DB, userID string, bookID, chapter int64) error {
	var numChapters int64
	if err := db.QueryRow(`
		SELECT COALESCE(numChapters, 0)
		  FROM books
		 WHERE id = ? AND user_id = ?;
	`, bookID, userID).Scan(&numChapters); err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	}
	if chapter < 1 || (numChapters > 0 && chapter > numChapters) {
		// Out of bounds (treat 0/NULL numChapters as "no chapters")
		return sql.ErrNoRows
	}

	_, err := db.Exec(`
		INSERT OR IGNORE INTO progress (book_id, chapter_index)
		VALUES (?, ?);
	`, bookID, chapter)
	return err
}

// RemoveProgress unmarks a single chapter for a book owned by user.
// Deleting a non-existent row is a no-op.
func RemoveProgress(db *sql.DB, userID string, bookID, chapter int64) error {
	var numChapters int64
	if err := db.QueryRow(`
		SELECT COALESCE(numChapters, 0)
		  FROM books
		 WHERE id = ? AND user_id = ?;
	`, bookID, userID).Scan(&numChapters); err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	}
	if chapter < 1 || (numChapters > 0 && chapter > numChapters) {
		return sql.ErrNoRows
	}

	_, err := db.Exec(`
		DELETE FROM progress
		 WHERE book_id = ? AND chapter_index = ?;
	`, bookID, chapter)
	return err
}
