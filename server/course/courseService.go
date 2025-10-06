package course

import (
	"database/sql"
	"errors"
	"strings"
)

type Course struct {
	ID           int64  `json:"id"`
	UniversityID string `json:"universityId"`
	Year         int64  `json:"year"`
	Term         int64  `json:"term"` // 1..4
	Code         string `json:"code"`
	Name         string `json:"name"`
}

// AddCourse inserts a new course for a university.
// Validates basic inputs, ensures university exists, and relies on the UNIQUE constraint
// (university_id, year, term, code) for conflict handling.
func AddCourse(db *sql.DB, universityID string, year, term int64, code, name string) (Course, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)

	if universityID == "" || year <= 0 || term < 1 || term > 4 || code == "" || name == "" {
		return Course{}, errors.New("invalid input")
	}

	// Ensure the university exists.
	var exists string
	if err := db.QueryRow(`SELECT id FROM universities WHERE id = ?`, universityID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return Course{}, errors.New("university not found")
		}
		return Course{}, err
	}

	// Insert the course.
	res, err := db.Exec(`
		INSERT INTO courses (university_id, year, term, code, name)
		VALUES (?, ?, ?, ?, ?)
	`, universityID, year, term, code, name)
	if err != nil {
		// Map UNIQUE violation to a friendly error.
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return Course{}, errors.New("course already exists")
		}
		return Course{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Course{}, err
	}

	var c Course
	if err := db.QueryRow(`
		SELECT id, university_id, year, term, code, name
		  FROM courses
		 WHERE id = ?
	`, id).Scan(&c.ID, &c.UniversityID, &c.Year, &c.Term, &c.Code, &c.Name); err != nil {
		return Course{}, err
	}
	return c, nil
}


// ListMyCoursesByUniversity returns the courses the given user is enrolled in
// for a specific university.
func ListMyCoursesByUniversity(db *sql.DB, userID, universityID string) ([]Course, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(universityID) == "" {
		return []Course{}, nil
	}

	rows, err := db.Query(`
		SELECT c.id, c.university_id, c.year, c.term, c.code, c.name
		  FROM user_courses uc
		  JOIN courses c ON c.id = uc.course_id
		 WHERE uc.user_id = ?
		   AND c.university_id = ?
		 ORDER BY c.year DESC, c.term DESC, c.code ASC
	`, userID, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Course, 0, 32)
	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.UniversityID, &c.Year, &c.Term, &c.Code, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
