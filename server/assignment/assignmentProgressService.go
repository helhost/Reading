package assignment

import (
	"database/sql"
	"errors"
)

// SetAssignmentProgress marks an assignment as completed (true) or not completed (false) for userID.
// Returns sql.ErrNoRows if the assignment doesn't exist.
func SetAssignmentProgress(db *sql.DB, userID string, assignmentID int64, completed bool) error {
	if userID == "" || assignmentID <= 0 {
		return errors.New("invalid input")
	}

	// Ensure assignment exists (404 semantics for callers).
	var exists int64
	if err := db.QueryRow(`SELECT id FROM assignments WHERE id = ?`, assignmentID).Scan(&exists); err != nil {
		return err // may be sql.ErrNoRows
	}

	if completed {
		// Upsert: create or set completed=1
		_, err := db.Exec(`
			INSERT INTO progress (user_id, assignment_id, completed)
			VALUES (?, ?, 1)
			ON CONFLICT(user_id, assignment_id) DO UPDATE SET completed = 1
		`, userID, assignmentID)
		return err
	}

	// Not completed ⇒ delete row (your preference)
	_, err := db.Exec(`
		DELETE FROM progress
		 WHERE user_id = ? AND assignment_id IS NOT NULL AND assignment_id = ?
	`, userID, assignmentID)
	return err
}
