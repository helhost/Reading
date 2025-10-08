package store

import (
	"database/sql"
	"fmt"
)

// execer is satisfied by *sql.DB and *sql.Tx so we can reuse the same functions in a tx.
type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// EnsureCalendar runs all calendar-related schema (table + triggers) in a single tx.
func EnsureCalendar(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin calendar tx: %w", err)
	}
	if err := EnsureCalendarIndex(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := EnsureCalendarTriggers(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit calendar tx: %w", err)
	}
	return nil
}

// EnsureCalendarIndex creates the calendar_index table and indexes (no-op if already present).
func EnsureCalendarIndex(db execer) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS calendar_index (
		uid TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		kind TEXT NOT NULL,
		source_id INTEGER NOT NULL,
		summary TEXT NOT NULL,
		deadline_epoch INTEGER,
		completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0,1)),
		last_modified_epoch INTEGER NOT NULL,
		seq INTEGER NOT NULL DEFAULT 0,
		cancelled_at INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_cal_user              ON calendar_index(user_id);
	CREATE INDEX IF NOT EXISTS idx_cal_deadline          ON calendar_index(deadline_epoch);
	CREATE INDEX IF NOT EXISTS idx_cal_kind_source       ON calendar_index(kind, source_id);
	CREATE INDEX IF NOT EXISTS idx_cal_user_deadline     ON calendar_index(user_id, deadline_epoch);
	CREATE INDEX IF NOT EXISTS idx_cal_kind_source_user  ON calendar_index(kind, source_id, user_id);

	CREATE TABLE IF NOT EXISTS calendar_tokens (
		token TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
		last_used_at INTEGER
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_cal_tok_user ON calendar_tokens(user_id);
	`)
	if err != nil {
		return fmt.Errorf("ensure calendar_index: %w", err)
	}
	return nil
}

// EnsureCalendarTriggers installs triggers that keep calendar_index in sync.
func EnsureCalendarTriggers(db execer) error {
	_, err := db.Exec(`
	/* ============ ASSIGNMENTS ============ */

	CREATE TRIGGER IF NOT EXISTS cal_uc_ins_assignments
	AFTER INSERT ON user_courses
	BEGIN
		INSERT INTO calendar_index
			(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
		SELECT
			'yourapp:assignment:' || a.id || ':user:' || NEW.user_id,
			NEW.user_id,
			'assignment',
			a.id,
			printf('[%s] %s — %s',
			       (SELECT code FROM courses WHERE id = a.course_id),
			       (SELECT name FROM courses WHERE id = a.course_id),
			       a.title),
			a.deadline,
			0,
			strftime('%s','now'),
			0,
			NULL
		FROM assignments a
		WHERE a.course_id = NEW.course_id
		ON CONFLICT(uid) DO UPDATE SET
			summary = excluded.summary,
			deadline_epoch = excluded.deadline_epoch,
			last_modified_epoch = strftime('%s','now'),
			seq = calendar_index.seq + 1,
			cancelled_at = NULL;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_uc_del_assignments
	AFTER DELETE ON user_courses
	BEGIN
		UPDATE calendar_index
		SET cancelled_at = strftime('%s','now'),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'assignment'
		  AND user_id = OLD.user_id
		  AND cancelled_at IS NULL
		  AND source_id IN (SELECT id FROM assignments WHERE course_id = OLD.course_id);
	END;

	CREATE TRIGGER IF NOT EXISTS cal_assign_ins
	AFTER INSERT ON assignments
	BEGIN
		INSERT INTO calendar_index
			(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
		SELECT
			'yourapp:assignment:' || NEW.id || ':user:' || uc.user_id,
			uc.user_id,
			'assignment',
			NEW.id,
			printf('[%s] %s — %s',
			       (SELECT code FROM courses WHERE id = NEW.course_id),
			       (SELECT name FROM courses WHERE id = NEW.course_id),
			       NEW.title),
			NEW.deadline,
			0,
			strftime('%s','now'),
			0,
			NULL
		FROM user_courses uc
		WHERE uc.course_id = NEW.course_id
		ON CONFLICT(uid) DO UPDATE SET
			summary = excluded.summary,
			deadline_epoch = excluded.deadline_epoch,
			last_modified_epoch = strftime('%s','now'),
			seq = calendar_index.seq + 1,
			cancelled_at = NULL;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_assign_upd_fields
	AFTER UPDATE OF title, deadline ON assignments
	BEGIN
		UPDATE calendar_index
		SET summary = printf('[%s] %s — %s',
		                     (SELECT code FROM courses WHERE id = NEW.course_id),
		                     (SELECT name FROM courses WHERE id = NEW.course_id),
		                     NEW.title),
		    deadline_epoch = NEW.deadline,
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'assignment' AND source_id = NEW.id;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_assign_del
	AFTER DELETE ON assignments
	BEGIN
		UPDATE calendar_index
		SET cancelled_at = strftime('%s','now'),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'assignment'
		  AND source_id = OLD.id
		  AND cancelled_at IS NULL;
	END;

	/* propagate course name/code edits into assignment summaries */
	CREATE TRIGGER IF NOT EXISTS cal_course_upd_name_code
	AFTER UPDATE OF name, code ON courses
	BEGIN
		UPDATE calendar_index
		SET summary = printf('[%s] %s — %s',
		                     NEW.code, NEW.name,
		                     (SELECT title FROM assignments WHERE id = calendar_index.source_id)),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'assignment'
		  AND source_id IN (SELECT id FROM assignments WHERE course_id = NEW.id);
	END;

	/* ============ ARTICLES ============ */

	CREATE TRIGGER IF NOT EXISTS cal_uc_ins_articles
	AFTER INSERT ON user_courses
	BEGIN
		INSERT INTO calendar_index
			(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
		SELECT
			'yourapp:article:' || ar.id || ':user:' || NEW.user_id,
			NEW.user_id,
			'article',
			ar.id,
			ar.title,
			ar.deadline,
			0,
			strftime('%s','now'),
			0,
			NULL
		FROM articles ar
		WHERE ar.course_id = NEW.course_id
		ON CONFLICT(uid) DO UPDATE SET
			summary = excluded.summary,
			deadline_epoch = excluded.deadline_epoch,
			last_modified_epoch = strftime('%s','now'),
			seq = calendar_index.seq + 1,
			cancelled_at = NULL;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_uc_del_articles
	AFTER DELETE ON user_courses
	BEGIN
		UPDATE calendar_index
		SET cancelled_at = strftime('%s','now'),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'article'
		  AND user_id = OLD.user_id
		  AND cancelled_at IS NULL
		  AND source_id IN (SELECT id FROM articles WHERE course_id = OLD.course_id);
	END;

	CREATE TRIGGER IF NOT EXISTS cal_article_ins
	AFTER INSERT ON articles
	BEGIN
		INSERT INTO calendar_index
			(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
		SELECT
			'yourapp:article:' || NEW.id || ':user:' || uc.user_id,
			uc.user_id,
			'article',
			NEW.id,
			NEW.title,
			NEW.deadline,
			0,
			strftime('%s','now'),
			0,
			NULL
		FROM user_courses uc
		WHERE uc.course_id = NEW.course_id
		ON CONFLICT(uid) DO UPDATE SET
			summary = excluded.summary,
			deadline_epoch = excluded.deadline_epoch,
			last_modified_epoch = strftime('%s','now'),
			seq = calendar_index.seq + 1,
			cancelled_at = NULL;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_article_upd_fields
	AFTER UPDATE OF title, deadline ON articles
	BEGIN
		UPDATE calendar_index
		SET summary = NEW.title,
		    deadline_epoch = NEW.deadline,
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'article' AND source_id = NEW.id;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_article_del
	AFTER DELETE ON articles
	BEGIN
		UPDATE calendar_index
		SET cancelled_at = strftime('%s','now'),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'article'
		  AND source_id = OLD.id
		  AND cancelled_at IS NULL;
	END;

	/* ============ CHAPTERS ============ */

	CREATE TRIGGER IF NOT EXISTS cal_uc_ins_chapters
	AFTER INSERT ON user_courses
	BEGIN
		INSERT INTO calendar_index
			(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
		SELECT
			'yourapp:chapter:' || c.id || ':user:' || NEW.user_id,
			NEW.user_id,
			'chapter',
			c.id,
			printf('Chapter %d — %s', c.chapter_num, b.title),
			c.deadline,
			0,
			strftime('%s','now'),
			0,
			NULL
		FROM chapters c
		JOIN books b ON b.id = c.book_id
		WHERE b.course_id = NEW.course_id
		ON CONFLICT(uid) DO UPDATE SET
			summary = excluded.summary,
			deadline_epoch = excluded.deadline_epoch,
			last_modified_epoch = strftime('%s','now'),
			seq = calendar_index.seq + 1,
			cancelled_at = NULL;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_uc_del_chapters
	AFTER DELETE ON user_courses
	BEGIN
		UPDATE calendar_index
		SET cancelled_at = strftime('%s','now'),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'chapter'
		  AND user_id = OLD.user_id
		  AND cancelled_at IS NULL
		  AND source_id IN (
			SELECT c.id
			FROM chapters c
			JOIN books b ON b.id = c.book_id
			WHERE b.course_id = OLD.course_id
		  );
	END;

	CREATE TRIGGER IF NOT EXISTS cal_chapter_ins
	AFTER INSERT ON chapters
	BEGIN
		INSERT INTO calendar_index
			(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
		SELECT
			'yourapp:chapter:' || NEW.id || ':user:' || uc.user_id,
			uc.user_id,
			'chapter',
			NEW.id,
			printf('Chapter %d — %s', NEW.chapter_num,
			       (SELECT title FROM books WHERE id = NEW.book_id)),
			NEW.deadline,
			0,
			strftime('%s','now'),
			0,
			NULL
		FROM user_courses uc
		WHERE uc.course_id = (SELECT course_id FROM books WHERE id = NEW.book_id)
		ON CONFLICT(uid) DO UPDATE SET
			summary = excluded.summary,
			deadline_epoch = excluded.deadline_epoch,
			last_modified_epoch = strftime('%s','now'),
			seq = calendar_index.seq + 1,
			cancelled_at = NULL;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_chapter_upd_fields
	AFTER UPDATE OF chapter_num, deadline ON chapters
	BEGIN
		UPDATE calendar_index
		SET summary = printf('Chapter %d — %s', NEW.chapter_num,
		                     (SELECT title FROM books WHERE id = NEW.book_id)),
		    deadline_epoch = NEW.deadline,
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'chapter' AND source_id = NEW.id;
	END;

	CREATE TRIGGER IF NOT EXISTS cal_book_upd_title
	AFTER UPDATE OF title ON books
	BEGIN
		UPDATE calendar_index
		SET summary = (
			SELECT printf('Chapter %d — %s', c.chapter_num, NEW.title)
			FROM chapters c
			WHERE c.id = calendar_index.source_id
		),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'chapter'
		  AND source_id IN (SELECT id FROM chapters WHERE book_id = NEW.id);
	END;

	CREATE TRIGGER IF NOT EXISTS cal_chapter_del
	AFTER DELETE ON chapters
	BEGIN
		UPDATE calendar_index
		SET cancelled_at = strftime('%s','now'),
		    last_modified_epoch = strftime('%s','now'),
		    seq = seq + 1
		WHERE kind = 'chapter'
		  AND source_id = OLD.id
		  AND cancelled_at IS NULL;
	END;
	`)
	if err != nil {
		return fmt.Errorf("ensure calendar triggers: %w", err)
	}
	return nil
}
