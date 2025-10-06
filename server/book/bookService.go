
package book

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"example.com/sqlite-server/chapter"
)

type Book struct {
	ID          int64              `json:"id"`
	CourseID    int64              `json:"courseId"`
	Title       string             `json:"title"`
	Author      string             `json:"author"`
	NumChapters *int64             `json:"numChapters,omitempty"`
	Location    *string            `json:"location,omitempty"`
	Chapters    []chapter.Chapter  `json:"chapters,omitempty"`
}

// AddBook inserts a new book and, if numChapters > 0, creates chapters [1..n] atomically.
// Returns the created book including its chapters.
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

	// Load chapters (still within TX for consistency)
	chaps, err := chapter.ListByBook(db, bookID)
	if err != nil {
		return Book{}, err
	}
	b.Chapters = chaps

	return b, nil
}

// ListBooksByCourse returns all books for a course including chapters (via chapter service).
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

	// Batch load chapters via chapter service
	chapterMap, err := chapter.ListByBooks(db, ids)
	if err != nil {
		return nil, err
	}

	// Attach
	for i := range books {
		books[i].Chapters = chapterMap[books[i].ID]
	}
	return books, nil
}
