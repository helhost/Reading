package assignment

import (
	"database/sql"
	"errors"
)

// AssignmentUniversityID returns the university_id for the assignment's course.
// If the assignment doesn't exist, returns sql.ErrNoRows.
func AssignmentUniversityID(db *sql.DB, assignmentID int64) (string, error) {
	if assignmentID <= 0 {
		return "", errors.New("invalid input")
	}
	var uniID string
	err := db.QueryRow(`
		SELECT c.university_id
		  FROM assignments a
		  JOIN courses c ON c.id = a.course_id
		 WHERE a.id = ?;
	`, assignmentID).Scan(&uniID)
	if err != nil {
		return "", err
	}
	return uniID, nil
}

// UserEnrolledInAssignmentCourse returns true if the given user has an enrollment
// (user_courses) in the course to which the assignment belongs.
func UserEnrolledInAssignmentCourse(db *sql.DB, userID string, assignmentID int64) (bool, error) {
	if userID == "" || assignmentID <= 0 {
		return false, errors.New("invalid input")
	}
	var exists int
	err := db.QueryRow(`
		SELECT 1
		  FROM user_courses uc
		  JOIN assignments a ON a.id = ?
		  JOIN courses c     ON c.id = a.course_id
		 WHERE uc.user_id = ? AND uc.course_id = c.id
		 LIMIT 1;
	`, assignmentID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists == 1, nil
}

// SetAssignmentDeadline sets assignments.deadline to the provided unix seconds,
// or clears it when deadline == nil. Returns sql.ErrNoRows if assignment not found.
func SetAssignmentDeadline(db *sql.DB, assignmentID int64, deadline *int64) error {
	if assignmentID <= 0 {
		return errors.New("invalid input")
	}
	res, err := db.Exec(`UPDATE assignments SET deadline = ? WHERE id = ?;`, deadline, assignmentID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
