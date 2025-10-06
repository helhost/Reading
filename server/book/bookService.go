
package book

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"example.com/sqlite-server/chapter"
)

type Book struct {
	ID          int64                             `json:"id"`
	CourseID    int64                             `json:"courseId"`
	Title       string                            `json:"title"`
	Author      string                            `json:"author"`
	NumChapters *int64                            `json:"numChapters,omitempty"`
	Location    *string                           `json:"location,omitempty"`
	Chapters    []chapter.ChapterWithStatus       `json:"chapters,omitempty"`
}

// AddBook inserts a new book and, if numChapters > 0, creates chapters [1..n] atomically.
// Returns the created book including its chapters (with Completed=false by default).
func AddBook(db *sql.DB, courseID int64, title, author string, numChapters *int64, location *string) (Book, error) {
	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if courseID <= 0 || title == "" || author == "" {
		return Book{}, errors.New("invalid input")
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return Book{}, err
	}
	defer func() { _ = tx.Rollback() }()

	// Ensure course exists within TX
	var cid int64
	if err := tx.QueryRow(`SELECT id FROM courses WHERE id = ?`, courseID).Scan(&cid); err != nil {
		if err == sql.ErrNoRows {
			return Book{}, errors.New("course not found")
		}
		return Book{}, err
	}

	// Insert book
	if _, err := tx.Exec(`
		INSERT INTO books (course_id, title, author, numChapters, location)
		VALUES (?, ?, ?, ?, ?)
	`, courseID, title, author, numChapters, location); err != nil {
		return Book{}, err
	}

	var bookID int64
	if err := tx.QueryRow(`SELECT last_insert_rowid();`).Scan(&bookID); err != nil {
		return Book{}, err
	}

	// Auto-create chapters if requested
	if numChapters != nil && *numChapters > 0 {
		if err := chapter.CreateChaptersRangeTx(tx, bookID, *numChapters); err != nil {
			return Book{}, err
		}
	}

	// Load created book inside TX
	var b Book
	if err := tx.QueryRow(`
		SELECT id, course_id, title, author, numChapters, location
		  FROM books
		 WHERE id = ?
	`, bookID).Scan(&b.ID, &b.CourseID, &b.Title, &b.Author, &b.NumChapters, &b.Location); err != nil {
		return Book{}, err
	}

	if err := tx.Commit(); err != nil {
		return Book{}, err
	}

	// Load chapters after commit and convert to ChapterWithStatus (Completed=false by default)
	chaps, err := chapter.ListByBook(db, bookID)
	if err != nil {
		return Book{}, err
	}
	ws := make([]chapter.ChapterWithStatus, 0, len(chaps))
	for _, c := range chaps {
		ws = append(ws, chapter.ChapterWithStatus{
			ID:         c.ID,
			BookID:     c.BookID,
			ChapterNum: c.ChapterNum,
			Deadline:   c.Deadline,
			Completed:  false,
		})
	}
	b.Chapters = ws

	return b, nil
}


// ListBooksByCourse returns all books for a course including chapters.
// Since Book.Chapters is []ChapterWithStatus, convert plain chapters to "with status" (Completed=false).
func ListBooksByCourse(db *sql.DB, courseID int64) ([]Book, error) {
	if courseID <= 0 {
		return []Book{}, nil
	}

	// Load books
	rows, err := db.Query(`
		SELECT id, course_id, title, author, numChapters, location
		  FROM books
		 WHERE course_id = ?
		 ORDER BY id ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]Book, 0, 32)
	ids := make([]int64, 0, 32)
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.CourseID, &b.Title, &b.Author, &b.NumChapters, &b.Location); err != nil {
			return nil, err
		}
		ids = append(ids, b.ID)
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return books, nil
	}

	// Batch load chapters via chapter service (plain chapters)
	chapterMap, err := chapter.ListByBooks(db, ids)
	if err != nil {
		return nil, err
	}

	// Attach (convert []chapter.Chapter -> []chapter.ChapterWithStatus)
	for i := range books {
		base := chapterMap[books[i].ID]
		ws := make([]chapter.ChapterWithStatus, 0, len(base))
		for _, c := range base {
			ws = append(ws, chapter.ChapterWithStatus{
				ID:         c.ID,
				BookID:     c.BookID,
				ChapterNum: c.ChapterNum,
				Deadline:   c.Deadline,
				Completed:  false,
			})
		}
		books[i].Chapters = ws
	}
	return books, nil
}


// ListBooksByCourseWithProgress returns all books for a course and attaches
// chapters annotated with the caller's Completed status.
func ListBooksByCourseWithProgress(db *sql.DB, courseID int64, userID string) ([]Book, error) {
	if courseID <= 0 || strings.TrimSpace(userID) == "" {
		return []Book{}, nil
	}

	// Load books
	rows, err := db.Query(`
		SELECT id, course_id, title, author, numChapters, location
		  FROM books
		 WHERE course_id = ?
		 ORDER BY id ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]Book, 0, 32)
	ids := make([]int64, 0, 32)
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.CourseID, &b.Title, &b.Author, &b.NumChapters, &b.Location); err != nil {
			return nil, err
		}
		ids = append(ids, b.ID)
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return books, nil
	}

	// Batch load chapters with per-user completion
	chapMap, err := chapter.ListByBooksWithProgress(db, ids, userID)
	if err != nil {
		return nil, err
	}

	for i := range books {
		books[i].Chapters = chapMap[books[i].ID]
	}
	return books, nil
}
