package article

import (
	"database/sql"
	"errors"
)

// SetArticleProgress marks an article as completed (true) or not completed (false) for userID.
// Returns sql.ErrNoRows if the article doesn't exist.
func SetArticleProgress(db *sql.DB, userID string, articleID int64, completed bool) error {
	if userID == "" || articleID <= 0 {
		return errors.New("invalid input")
	}

	// Ensure article exists (404 semantics for callers).
	var exists int64
	if err := db.QueryRow(`SELECT id FROM articles WHERE id = ?`, articleID).Scan(&exists); err != nil {
		return err // may be sql.ErrNoRows
	}

	if completed {
		// Upsert: create or set completed=1
		_, err := db.Exec(`
			INSERT INTO progress (user_id, article_id, completed)
			VALUES (?, ?, 1)
			ON CONFLICT(user_id, article_id) DO UPDATE SET completed = 1
		`, userID, articleID)
		return err
	}

	// Not completed → delete the row
	_, err := db.Exec(`
		DELETE FROM progress
		 WHERE user_id = ? AND article_id IS NOT NULL AND article_id = ?
	`, userID, articleID)
	return err
}
