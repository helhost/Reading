package chapter

import (
	"database/sql"
	"errors"
)

// ChapterBelongsToUniversity returns (true, nil) if the chapter exists and its course's
// university_id matches uniID. If the chapter doesn't exist, returns (false, sql.ErrNoRows).
func ChapterBelongsToUniversity(db *sql.DB, chapterID int64, uniID string) (bool, error) {
	if chapterID <= 0 || uniID == "" {
		return false, errors.New("invalid input")
	}
	var got string
	err := db.QueryRow(`
		SELECT c.university_id
		  FROM chapters ch
		  JOIN books b   ON b.id = ch.book_id
		  JOIN courses c ON c.id = b.course_id
		 WHERE ch.id = ?;
	`, chapterID).Scan(&got)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, sql.ErrNoRows
		}
		return false, err
	}
	return got == uniID, nil
}

// ChapterUniversityID returns the university_id for the chapter's course.
// If the chapter doesn't exist, returns sql.ErrNoRows.
func ChapterUniversityID(db *sql.DB, chapterID int64) (string, error) {
	if chapterID <= 0 {
		return "", errors.New("invalid input")
	}
	var uniID string
	err := db.QueryRow(`
		SELECT c.university_id
		  FROM chapters ch
		  JOIN books b   ON b.id = ch.book_id
		  JOIN courses c ON c.id = b.course_id
		 WHERE ch.id = ?;
	`, chapterID).Scan(&uniID)
	if err != nil {
		return "", err
	}
	return uniID, nil
}

// UserEnrolledInChapterCourse returns true if the given user has an enrollment
// (user_courses) in the course to which the chapter belongs.
func UserEnrolledInChapterCourse(db *sql.DB, userID string, chapterID int64) (bool, error) {
	if userID == "" || chapterID <= 0 {
		return false, errors.New("invalid input")
	}
	var exists int
	err := db.QueryRow(`
		SELECT 1
		  FROM user_courses uc
		  JOIN chapters ch ON ch.id = ?
		  JOIN books b     ON b.id = ch.book_id
		  JOIN courses c   ON c.id = b.course_id
		 WHERE uc.user_id = ? AND uc.course_id = c.id
		 LIMIT 1;
	`, chapterID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists == 1, nil
}

// SetChapterDeadline sets chapters.deadline to the provided unix seconds,
// or clears it when deadline == nil. Returns sql.ErrNoRows if chapter not found.
func SetChapterDeadline(db *sql.DB, chapterID int64, deadline *int64) error {
	if chapterID <= 0 {
		return errors.New("invalid input")
	}
	res, err := db.Exec(`UPDATE chapters SET deadline = ? WHERE id = ?;`, deadline, chapterID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
