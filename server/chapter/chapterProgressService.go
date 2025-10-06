package chapter

import (
	"database/sql"
	"errors"
)

// SetChapterProgress sets (completed=true) or clears (completed=false) the current user's
// progress row for a chapter. Returns sql.ErrNoRows if the chapter doesn't exist.
func SetChapterProgress(db *sql.DB, userID string, chapterID int64, completed bool) error {
	if userID == "" || chapterID <= 0 {
		return errors.New("invalid input")
	}

	// Ensure chapter exists (404 semantics).
	var exists int64
	if err := db.QueryRow(`SELECT id FROM chapters WHERE id = ?`, chapterID).Scan(&exists); err != nil {
		return err // may be sql.ErrNoRows
	}

	if completed {
		// Upsert: either create or set completed=1
		_, err := db.Exec(`
			INSERT INTO progress (user_id, chapter_id, completed)
			VALUES (?, ?, 1)
			ON CONFLICT(user_id, chapter_id) DO UPDATE SET completed = 1
		`, userID, chapterID)
		return err
	}

	// Not completed → delete row (your stated preference)
	_, err := db.Exec(`
		DELETE FROM progress
		 WHERE user_id = ? AND chapter_id IS NOT NULL AND chapter_id = ?
	`, userID, chapterID)
	return err
}


// ChapterCompleted returns whether the current user has a progress row for this chapter.
func ChapterCompleted(db *sql.DB, userID string, chapterID int64) (bool, error) {
	if userID == "" || chapterID <= 0 {
		return false, errors.New("invalid input")
	}
	var cnt int64
	if err := db.QueryRow(`
		SELECT COUNT(1)
		  FROM progress
		 WHERE user_id = ? AND chapter_id = ?
	`, userID, chapterID).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}
