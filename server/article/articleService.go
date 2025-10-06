package article

import (
	"database/sql"
	"errors"
	"strings"
)

type Article struct {
	ID        int64   `json:"id"`
	CourseID  int64   `json:"courseId"`
	Title     string  `json:"title"`
	Author    string  `json:"author"`
	Location  *string `json:"location,omitempty"`
	Deadline  *int64  `json:"deadline,omitempty"` // always nil on create
}

// AddArticle inserts a new article for a course and initializes deadline = NULL.
// Returns the created row.
func AddArticle(db *sql.DB, courseID int64, title, author string, location *string) (Article, error) {
	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if courseID <= 0 || title == "" || author == "" {
		return Article{}, errors.New("invalid input")
	}

	// Ensure the course exists.
	var cid int64
	if err := db.QueryRow(`SELECT id FROM courses WHERE id = ?`, courseID).Scan(&cid); err != nil {
		if err == sql.ErrNoRows {
			return Article{}, errors.New("course not found")
		}
		return Article{}, err
	}

	// Insert with deadline = NULL explicitly.
	if _, err := db.Exec(`
		INSERT INTO articles (course_id, title, author, location, deadline)
		VALUES (?, ?, ?, ?, NULL)
	`, courseID, title, author, location); err != nil {
		return Article{}, err
	}

	var id int64
	if err := db.QueryRow(`SELECT last_insert_rowid();`).Scan(&id); err != nil {
		return Article{}, err
	}

	var a Article
	var loc sql.NullString
	// deadline will be NULL on create
	if err := db.QueryRow(`
		SELECT id, course_id, title, author, location, deadline
		  FROM articles
		 WHERE id = ?
	`, id).Scan(&a.ID, &a.CourseID, &a.Title, &a.Author, &loc, new(sql.NullInt64)); err != nil {
		return Article{}, err
	}
	if loc.Valid {
		val := loc.String
		a.Location = &val
	}
	// a.Deadline remains nil
	return a, nil
}


// ListArticlesByCourse returns all articles for a course (including nullable deadline).
func ListArticlesByCourse(db *sql.DB, courseID int64) ([]Article, error) {
	if courseID <= 0 {
		return []Article{}, nil
	}

	rows, err := db.Query(`
		SELECT id, course_id, title, author, location, deadline
		  FROM articles
		 WHERE course_id = ?
		 ORDER BY id ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Article, 0, 32)
	for rows.Next() {
		var a Article
		var loc sql.NullString
		var dl  sql.NullInt64
		if err := rows.Scan(&a.ID, &a.CourseID, &a.Title, &a.Author, &loc, &dl); err != nil {
			return nil, err
		}
		if loc.Valid {
			v := loc.String
			a.Location = &v
		}
		if dl.Valid {
			v := dl.Int64
			a.Deadline = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetArticle returns the article by ID, including nullable location and deadline.
func GetArticle(db *sql.DB, id int64) (Article, error) {
	var a Article
	var loc sql.NullString
	var dl sql.NullInt64
	err := db.QueryRow(`
		SELECT id, course_id, title, author, location, deadline
		  FROM articles
		 WHERE id = ?;
	`, id).Scan(&a.ID, &a.CourseID, &a.Title, &a.Author, &loc, &dl)
	if err != nil {
		return Article{}, err
	}
	if loc.Valid {
		v := loc.String
		a.Location = &v
	}
	if dl.Valid {
		v := dl.Int64
		a.Deadline = &v
	}
	return a, nil
}
