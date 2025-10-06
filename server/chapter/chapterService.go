package chapter

import (
	"database/sql"
	"strings"
)

type Chapter struct {
	ID         int64  `json:"id"`
	BookID     int64  `json:"bookId"`
	ChapterNum int64  `json:"chapter_num"`
	// Optional later: *int64 `json:"deadline,omitempty"`
}

// CreateChaptersRangeTx inserts chapters 1..n for a book using an existing transaction.
func CreateChaptersRangeTx(tx *sql.Tx, bookID int64, n int64) error {
	if n <= 0 {
		return nil
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

// ListByBook returns all chapters for a single book.
func ListByBook(db *sql.DB, bookID int64) ([]Chapter, error) {
	rows, err := db.Query(`
		SELECT id, book_id, chapter_num
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
		if err := rows.Scan(&c.ID, &c.BookID, &c.ChapterNum); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListByBooks returns chapters for many books in one call: map[bookID][]Chapter.
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
		SELECT id, book_id, chapter_num
		  FROM chapters
		 WHERE book_id IN (` + strings.Join(placeholders, ",") + `)
		 ORDER BY book_id ASC, chapter_num ASC
	`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64][]Chapter, len(bookIDs))
	for rows.Next() {
		var c Chapter
		if err := rows.Scan(&c.ID, &c.BookID, &c.ChapterNum); err != nil {
			return nil, err
		}
		out[c.BookID] = append(out[c.BookID], c)
	}
	return out, rows.Err()
}
