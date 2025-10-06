package assignment

import (
	"database/sql"
	"errors"
	"strings"
)

type Assignment struct {
	ID          int64   `json:"id"`
	CourseID    int64   `json:"courseId"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Deadline    *int64  `json:"deadline,omitempty"`
}

// AddAssignment inserts a new assignment for a course with deadline = NULL.
func AddAssignment(db *sql.DB, courseID int64, title string, description *string) (Assignment, error) {
	title = strings.TrimSpace(title)
	if courseID <= 0 || title == "" {
		return Assignment{}, errors.New("invalid input")
	}
	if description != nil {
		s := strings.TrimSpace(*description)
		if s == "" {
			description = nil
		} else {
			description = &s
		}
	}

	// Ensure course exists
	var cid int64
	if err := db.QueryRow(`SELECT id FROM courses WHERE id = ?`, courseID).Scan(&cid); err != nil {
		if err == sql.ErrNoRows {
			return Assignment{}, errors.New("course not found")
		}
		return Assignment{}, err
	}

	// Insert with NULL deadline
	if _, err := db.Exec(`
		INSERT INTO assignments (course_id, title, description, deadline)
		VALUES (?, ?, ?, NULL)
	`, courseID, title, description); err != nil {
		return Assignment{}, err
	}

	var id int64
	if err := db.QueryRow(`SELECT last_insert_rowid()`).Scan(&id); err != nil {
		return Assignment{}, err
	}

	return GetAssignment(db, id)
}

// GetAssignment returns the assignment by ID.
func GetAssignment(db *sql.DB, id int64) (Assignment, error) {
	var a Assignment
	var desc sql.NullString
	var dl sql.NullInt64
	err := db.QueryRow(`
		SELECT id, course_id, title, description, deadline
		  FROM assignments
		 WHERE id = ?
	`, id).Scan(&a.ID, &a.CourseID, &a.Title, &desc, &dl)
	if err != nil {
		return Assignment{}, err
	}
	if desc.Valid {
		v := desc.String
		a.Description = &v
	}
	if dl.Valid {
		v := dl.Int64
		a.Deadline = &v
	}
	return a, nil
}

// ListAssignmentsByCourse returns all assignments for a course.
func ListAssignmentsByCourse(db *sql.DB, courseID int64) ([]Assignment, error) {
	if courseID <= 0 {
		return []Assignment{}, nil
	}

	rows, err := db.Query(`
		SELECT id, course_id, title, description, deadline
		  FROM assignments
		 WHERE course_id = ?
		 ORDER BY id ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Assignment, 0, 32)
	for rows.Next() {
		var a Assignment
		var desc sql.NullString
		var dl sql.NullInt64
		if err := rows.Scan(&a.ID, &a.CourseID, &a.Title, &desc, &dl); err != nil {
			return nil, err
		}
		if desc.Valid {
			v := desc.String
			a.Description = &v
		}
		if dl.Valid {
			v := dl.Int64
			a.Deadline = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}


type AssignmentWithStatus struct {
	ID          int64   `json:"id"`
	CourseID    int64   `json:"courseId"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Deadline    *int64  `json:"deadline,omitempty"`
	Completed   bool    `json:"completed"`
}

// ListAssignmentsByCourseWithProgress returns all assignments for a course
// and marks whether the given user has completed each one.
func ListAssignmentsByCourseWithProgress(db *sql.DB, courseID int64, userID string) ([]AssignmentWithStatus, error) {
	if courseID <= 0 || strings.TrimSpace(userID) == "" {
		return []AssignmentWithStatus{}, nil
	}

	rows, err := db.Query(`
		SELECT a.id, a.course_id, a.title, a.description, a.deadline,
		       COALESCE(p.completed, 0)
		  FROM assignments a
		  LEFT JOIN progress p
		         ON p.assignment_id = a.id
		        AND p.user_id = ?
		 WHERE a.course_id = ?
		 ORDER BY a.id ASC
	`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AssignmentWithStatus, 0, 32)
	for rows.Next() {
		var a AssignmentWithStatus
		var desc sql.NullString
		var dl sql.NullInt64
		var compInt int64
		if err := rows.Scan(&a.ID, &a.CourseID, &a.Title, &desc, &dl, &compInt); err != nil {
			return nil, err
		}
		if desc.Valid {
			v := desc.String
			a.Description = &v
		}
		if dl.Valid {
			v := dl.Int64
			a.Deadline = &v
		}
		a.Completed = compInt == 1
		out = append(out, a)
	}
	return out, rows.Err()
}
