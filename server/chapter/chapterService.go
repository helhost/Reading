package chapter

import (
	"database/sql"
	"errors"
	"strings"
)

type Chapter struct {
	ID         int64   `json:"id"`
	BookID     int64   `json:"bookId"`
	ChapterNum int64   `json:"chapter_num"`
	Deadline   *int64  `json:"deadline,omitempty"`
}

// CreateChaptersRangeTx inserts chapters 1..n for a book, inside the provided transaction.
func CreateChaptersRangeTx(tx *sql.Tx, bookID int64, n int64) error {
	if bookID <= 0 || n <= 0 {
		return errors.New("invalid input")
	}
	stmt, err := tx.Prepare(`INSERT INTO chapters (book_id, chapter_num) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := int64(1); i <= n; i++ {
		if _, err := stmt.Exec(bookID, i); err != nil {
			return err
		}
	}
	return nil
}

// UpdateDeadline sets or clears a deadline for a single chapter.
// deadline == nil clears it; otherwise sets to the provided unix seconds.
func UpdateDeadline(db *sql.DB, chapterID int64, deadline *int64) error {
	if chapterID <= 0 {
		return errors.New("invalid input")
	}
	res, err := db.Exec(`UPDATE chapters SET deadline = ? WHERE id = ?`, deadline, chapterID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListByBook returns all chapters for a given book, including deadline.
func ListByBook(db *sql.DB, bookID int64) ([]Chapter, error) {
	if bookID <= 0 {
		return []Chapter{}, nil
	}

	rows, err := db.Query(`
		SELECT id, book_id, chapter_num, deadline
		  FROM chapters
		 WHERE book_id = ?
		 ORDER BY chapter_num ASC
	`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Chapter, 0, 64)
	for rows.Next() {
		var c Chapter
		var dl sql.NullInt64
		if err := rows.Scan(&c.ID, &c.BookID, &c.ChapterNum, &dl); err != nil {
			return nil, err
		}
		if dl.Valid {
			v := dl.Int64
			c.Deadline = &v
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListByBooks returns chapters grouped by bookID for a set of book IDs, including deadline.
func ListByBooks(db *sql.DB, bookIDs []int64) (map[int64][]Chapter, error) {
	if len(bookIDs) == 0 {
		return map[int64][]Chapter{}, nil
	}

	placeholders := make([]string, 0, len(bookIDs))
	args := make([]any, 0, len(bookIDs))
	for _, id := range bookIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	q := `
		SELECT id, book_id, chapter_num, deadline
		  FROM chapters
		 WHERE book_id IN (` + strings.Join(placeholders, ",") + `)
		 ORDER BY book_id ASC, chapter_num ASC
	`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int64][]Chapter, len(bookIDs))
	for rows.Next() {
		var c Chapter
		var dl sql.NullInt64
		if err := rows.Scan(&c.ID, &c.BookID, &c.ChapterNum, &dl); err != nil {
			return nil, err
		}
		if dl.Valid {
			v := dl.Int64
			c.Deadline = &v
		}
		m[c.BookID] = append(m[c.BookID], c)
	}
	return m, rows.Err()
}

func GetChapter(db *sql.DB, id int64) (Chapter, error) {
	var c Chapter
	var dl sql.NullInt64
	err := db.QueryRow(`
		SELECT id, book_id, chapter_num, deadline
		  FROM chapters
		 WHERE id = ?
	`, id).Scan(&c.ID, &c.BookID, &c.ChapterNum, &dl)
	if err != nil {
		return Chapter{}, err
	}
	if dl.Valid {
		v := dl.Int64
		c.Deadline = &v
	} else {
		c.Deadline = nil
	}
	return c, nil
}


type ChapterWithStatus struct {
	ID         int64  `json:"id"`
	BookID     int64  `json:"bookId"`
	ChapterNum int64  `json:"chapter_num"`
	Deadline   *int64 `json:"deadline,omitempty"`
	Completed  bool   `json:"completed"`
}

// ListByBooksWithProgress returns chapters for the given book IDs and marks
// whether the given user has completed each one.
func ListByBooksWithProgress(db *sql.DB, bookIDs []int64, userID string) (map[int64][]ChapterWithStatus, error) {
	if len(bookIDs) == 0 || strings.TrimSpace(userID) == "" {
		return map[int64][]ChapterWithStatus{}, nil
	}

	placeholders := make([]string, 0, len(bookIDs))
	args := make([]any, 0, len(bookIDs)+1)
	args = append(args, userID)
	for _, id := range bookIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	q := `
		SELECT ch.id, ch.book_id, ch.chapter_num, ch.deadline,
		       COALESCE(p.completed, 0)
		  FROM chapters ch
		  LEFT JOIN progress p
		         ON p.chapter_id = ch.id
		        AND p.user_id = ?
		 WHERE ch.book_id IN (` + strings.Join(placeholders, ",") + `)
		 ORDER BY ch.book_id ASC, ch.chapter_num ASC
	`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int64][]ChapterWithStatus, len(bookIDs))
	for rows.Next() {
		var c ChapterWithStatus
		var dl sql.NullInt64
		var compInt int64
		if err := rows.Scan(&c.ID, &c.BookID, &c.ChapterNum, &dl, &compInt); err != nil {
			return nil, err
		}
		if dl.Valid {
			v := dl.Int64
			c.Deadline = &v
		}
		c.Completed = compInt == 1
		m[c.BookID] = append(m[c.BookID], c)
	}
	return m, rows.Err()
}
