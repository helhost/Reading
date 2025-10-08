package store

import (
	"database/sql"
	"fmt"
)

// One-off backfill for existing rows in calendar_index.
// Adds "[CODE] Course — …" to summaries for assignments, articles, and chapters.
func Backfill_Calendar_AllKindsIncludeCourse(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin backfill tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Assignments: "[CODE] Course — <assignment title>"
	if _, err := tx.Exec(`
UPDATE calendar_index
SET
  summary = printf('[%s] %s — %s',
    (SELECT c.code FROM assignments a JOIN courses c ON c.id = a.course_id WHERE a.id = calendar_index.source_id),
    (SELECT c.name FROM assignments a JOIN courses c ON c.id = a.course_id WHERE a.id = calendar_index.source_id),
    (SELECT a.title FROM assignments a WHERE a.id = calendar_index.source_id)
  ),
  last_modified_epoch = strftime('%s','now'),
  seq = seq + 1
WHERE kind = 'assignment'
  AND EXISTS (SELECT 1 FROM assignments a JOIN courses c ON c.id = a.course_id WHERE a.id = calendar_index.source_id);
`); err != nil {
		return fmt.Errorf("backfill assignments: %w", err)
	}

	// Articles: "[CODE] Course — <article title>"
	if _, err := tx.Exec(`
UPDATE calendar_index
SET
  summary = printf('[%s] %s — %s',
    (SELECT c.code FROM articles ar JOIN courses c ON c.id = ar.course_id WHERE ar.id = calendar_index.source_id),
    (SELECT c.name FROM articles ar JOIN courses c ON c.id = ar.course_id WHERE ar.id = calendar_index.source_id),
    (SELECT ar.title FROM articles ar WHERE ar.id = calendar_index.source_id)
  ),
  last_modified_epoch = strftime('%s','now'),
  seq = seq + 1
WHERE kind = 'article'
  AND EXISTS (SELECT 1 FROM articles ar JOIN courses c ON c.id = ar.course_id WHERE ar.id = calendar_index.source_id);
`); err != nil {
		return fmt.Errorf("backfill articles: %w", err)
	}

	// Chapters: "[CODE] Course — Chapter N — <book title>"
	if _, err := tx.Exec(`
UPDATE calendar_index
SET
  summary = printf('[%s] %s — Chapter %d — %s',
    (SELECT c.code FROM chapters ch JOIN books b ON b.id = ch.book_id JOIN courses c ON c.id = b.course_id WHERE ch.id = calendar_index.source_id),
    (SELECT c.name FROM chapters ch JOIN books b ON b.id = ch.book_id JOIN courses c ON c.id = b.course_id WHERE ch.id = calendar_index.source_id),
    (SELECT ch.chapter_num FROM chapters ch WHERE ch.id = calendar_index.source_id),
    (SELECT b.title FROM chapters ch JOIN books b ON b.id = ch.book_id WHERE ch.id = calendar_index.source_id)
  ),
  last_modified_epoch = strftime('%s','now'),
  seq = seq + 1
WHERE kind = 'chapter'
  AND EXISTS (
    SELECT 1 FROM chapters ch
    JOIN books b ON b.id = ch.book_id
    JOIN courses c ON c.id = b.course_id
    WHERE ch.id = calendar_index.source_id
  );
`); err != nil {
		return fmt.Errorf("backfill chapters: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit backfill tx: %w", err)
	}
	return nil
}
