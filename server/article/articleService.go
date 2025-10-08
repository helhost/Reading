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


type ArticleWithStatus struct {
	ID        int64   `json:"id"`
	CourseID  int64   `json:"courseId"`
	Title     string  `json:"title"`
	Author    string  `json:"author"`
	Location  *string `json:"location,omitempty"`
	Deadline  *int64  `json:"deadline,omitempty"`
	Completed bool    `json:"completed"`
}

// ListArticlesByCourseWithProgress returns all articles for a course and
// includes a per-user "completed" flag from the progress table.
func ListArticlesByCourseWithProgress(db *sql.DB, courseID int64, userID string) ([]ArticleWithStatus, error) {
	if courseID <= 0 || strings.TrimSpace(userID) == "" {
		return []ArticleWithStatus{}, nil
	}

	rows, err := db.Query(`
		SELECT a.id, a.course_id, a.title, a.author, a.location, a.deadline,
		       COALESCE(p.completed, 0)
		  FROM articles a
		  LEFT JOIN progress p
		         ON p.article_id = a.id
		        AND p.user_id   = ?
		 WHERE a.course_id = ?
		 ORDER BY a.id ASC
	`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ArticleWithStatus, 0, 32)
	for rows.Next() {
		var a ArticleWithStatus
		var loc sql.NullString
		var dl sql.NullInt64
		var compInt int64
		if err := rows.Scan(&a.ID, &a.CourseID, &a.Title, &a.Author, &loc, &dl, &compInt); err != nil {
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
		a.Completed = compInt == 1
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


// DeleteArticleIfNoProgress deletes the article iff it exists and
// no user has a completion row for it. Returns (false, sql.ErrNoRows) if missing.
func DeleteArticleIfNoProgress(db *sql.DB, articleID int64) (bool, error) {
	if articleID <= 0 {
		return false, errors.New("invalid input")
	}

	// Ensure it exists
	var exists int64
	if err := db.QueryRow(`SELECT id FROM articles WHERE id = ?`, articleID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, sql.ErrNoRows
		}
		return false, err
	}

	// Block if anyone has progress
	var cnt int64
	if err := db.QueryRow(`SELECT COUNT(1) FROM progress WHERE article_id = ?`, articleID).Scan(&cnt); err != nil {
		return false, err
	}
	if cnt > 0 {
		return false, errors.New("article has progress")
	}

	// Delete
	res, err := db.Exec(`DELETE FROM articles WHERE id = ?`, articleID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
