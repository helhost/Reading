
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
	Term         int64  `json:"term"`
	Code         string `json:"code"`
	Name         string `json:"name"`
}

// AddCourse inserts a new course for a university.
// Enforces: university exists; sane year/term; unique per (university_id, year, term, code).
func AddCourse(db *sql.DB, universityID string, year, term int64, code, name string) (Course, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if universityID == "" || year <= 0 || term < 1 || term > 4 || code == "" || name == "" {
		return Course{}, errors.New("invalid input")
	}

	// Ensure university exists
	var tmp string
	if err := db.QueryRow(`SELECT id FROM universities WHERE id = ?`, universityID).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return Course{}, errors.New("university not found")
		}
		return Course{}, err
	}

	// Insert
	res, err := db.Exec(`
		INSERT INTO courses (university_id, year, term, code, name)
		VALUES (?, ?, ?, ?, ?)
	`, universityID, year, term, code, name)
	if err != nil {
		// Rely on UNIQUE constraint for conflicts
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
