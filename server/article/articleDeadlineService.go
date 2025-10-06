package article

import (
	"database/sql"
	"errors"
)

// ArticleBelongsToUniversity returns (true,nil) if the article exists and its course's
// university_id matches uniID. If the article doesn't exist, returns (false, sql.ErrNoRows).
func ArticleBelongsToUniversity(db *sql.DB, articleID int64, uniID string) (bool, error) {
	if articleID <= 0 || uniID == "" {
		return false, errors.New("invalid input")
	}
	var got string
	err := db.QueryRow(`
		SELECT c.university_id
		  FROM articles a
		  JOIN courses  c ON c.id = a.course_id
		 WHERE a.id = ?;
	`, articleID).Scan(&got)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, sql.ErrNoRows
		}
		return false, err
	}
	return got == uniID, nil
}

// ArticleUniversityID returns the university_id for the article's course.
// If the article doesn't exist, returns sql.ErrNoRows.
func ArticleUniversityID(db *sql.DB, articleID int64) (string, error) {
	if articleID <= 0 {
		return "", errors.New("invalid input")
	}
	var uniID string
	err := db.QueryRow(`
		SELECT c.university_id
		  FROM articles a
		  JOIN courses  c ON c.id = a.course_id
		 WHERE a.id = ?;
	`, articleID).Scan(&uniID)
	if err != nil {
		return "", err
	}
	return uniID, nil
}

// UserEnrolledInArticleCourse returns true if the given user has an enrollment
// (user_courses) in the course to which the article belongs.
func UserEnrolledInArticleCourse(db *sql.DB, userID string, articleID int64) (bool, error) {
	if userID == "" || articleID <= 0 {
		return false, errors.New("invalid input")
	}
	var exists int
	err := db.QueryRow(`
		SELECT 1
		  FROM user_courses uc
		  JOIN articles a ON a.id = ?
		  JOIN courses  c ON c.id = a.course_id
		 WHERE uc.user_id = ? AND uc.course_id = c.id
		 LIMIT 1;
	`, articleID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists == 1, nil
}

// SetArticleDeadline sets articles.deadline to the provided unix seconds,
// or clears it when deadline == nil. Returns sql.ErrNoRows if article not found.
func SetArticleDeadline(db *sql.DB, articleID int64, deadline *int64) error {
	if articleID <= 0 {
		return errors.New("invalid input")
	}
	res, err := db.Exec(`UPDATE articles SET deadline = ? WHERE id = ?;`, deadline, articleID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
