package store

import (
	"database/sql"
	"fmt"
)

// One-off migration: ensure all calendar summaries emitted by triggers
// include "[CODE] Course — …" for assignments, articles, and chapters.
// Safe to run multiple times; only affects triggers.
func Migrate_Calendar_AllKinds_IncludeCourse(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`
/* =======================
   ASSIGNMENTS
   ======================= */
DROP TRIGGER IF EXISTS cal_uc_ins_assignments;
CREATE TRIGGER cal_uc_ins_assignments
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

DROP TRIGGER IF EXISTS cal_assign_ins;
CREATE TRIGGER cal_assign_ins
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

DROP TRIGGER IF EXISTS cal_assign_upd_fields;
CREATE TRIGGER cal_assign_upd_fields
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

DROP TRIGGER IF EXISTS cal_course_upd_name_code;
CREATE TRIGGER cal_course_upd_name_code
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

/* =======================
   ARTICLES
   ======================= */
DROP TRIGGER IF EXISTS cal_uc_ins_articles;
CREATE TRIGGER cal_uc_ins_articles
AFTER INSERT ON user_courses
BEGIN
	INSERT INTO calendar_index
		(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
	SELECT
		'yourapp:article:' || ar.id || ':user:' || NEW.user_id,
		NEW.user_id,
		'article',
		ar.id,
		printf('[%s] %s — %s',
		       (SELECT code FROM courses WHERE id = ar.course_id),
		       (SELECT name FROM courses WHERE id = ar.course_id),
		       ar.title),
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

DROP TRIGGER IF EXISTS cal_article_ins;
CREATE TRIGGER cal_article_ins
AFTER INSERT ON articles
BEGIN
	INSERT INTO calendar_index
		(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
	SELECT
		'yourapp:article:' || NEW.id || ':user:' || uc.user_id,
		uc.user_id,
		'article',
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

DROP TRIGGER IF EXISTS cal_article_upd_fields;
CREATE TRIGGER cal_article_upd_fields
AFTER UPDATE OF title, deadline ON articles
BEGIN
	UPDATE calendar_index
	SET summary = printf('[%s] %s — %s',
	                     (SELECT code FROM courses WHERE id = NEW.course_id),
	                     (SELECT name FROM courses WHERE id = NEW.course_id),
	                     NEW.title),
	    deadline_epoch = NEW.deadline,
	    last_modified_epoch = strftime('%s','now'),
	    seq = seq + 1
	WHERE kind = 'article' AND source_id = NEW.id;
END;

DROP TRIGGER IF EXISTS cal_course_upd_name_code_articles;
CREATE TRIGGER cal_course_upd_name_code_articles
AFTER UPDATE OF name, code ON courses
BEGIN
	UPDATE calendar_index
	SET summary = printf('[%s] %s — %s',
	                     NEW.code, NEW.name,
	                     (SELECT title FROM articles WHERE id = calendar_index.source_id)),
	    last_modified_epoch = strftime('%s','now'),
	    seq = seq + 1
	WHERE kind = 'article'
	  AND source_id IN (SELECT id FROM articles WHERE course_id = NEW.id);
END;

/* =======================
   CHAPTERS
   ======================= */
DROP TRIGGER IF EXISTS cal_uc_ins_chapters;
CREATE TRIGGER cal_uc_ins_chapters
AFTER INSERT ON user_courses
BEGIN
	INSERT INTO calendar_index
		(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
	SELECT
		'yourapp:chapter:' || c.id || ':user:' || NEW.user_id,
		NEW.user_id,
		'chapter',
		c.id,
		printf('[%s] %s — Chapter %d — %s',
		       (SELECT code FROM courses WHERE id = b.course_id),
		       (SELECT name FROM courses WHERE id = b.course_id),
		       c.chapter_num,
		       b.title),
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

DROP TRIGGER IF EXISTS cal_chapter_ins;
CREATE TRIGGER cal_chapter_ins
AFTER INSERT ON chapters
BEGIN
	INSERT INTO calendar_index
		(uid, user_id, kind, source_id, summary, deadline_epoch, completed, last_modified_epoch, seq, cancelled_at)
	SELECT
		'yourapp:chapter:' || NEW.id || ':user:' || uc.user_id,
		uc.user_id,
		'chapter',
		NEW.id,
		printf('[%s] %s — Chapter %d — %s',
		       (SELECT code FROM courses WHERE id = (SELECT course_id FROM books WHERE id = NEW.book_id)),
		       (SELECT name FROM courses WHERE id = (SELECT course_id FROM books WHERE id = NEW.book_id)),
		       NEW.chapter_num,
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

DROP TRIGGER IF EXISTS cal_chapter_upd_fields;
CREATE TRIGGER cal_chapter_upd_fields
AFTER UPDATE OF chapter_num, deadline ON chapters
BEGIN
	UPDATE calendar_index
	SET summary = printf('[%s] %s — Chapter %d — %s',
	                     (SELECT code FROM courses WHERE id = (SELECT course_id FROM books WHERE id = NEW.book_id)),
	                     (SELECT name FROM courses WHERE id = (SELECT course_id FROM books WHERE id = NEW.book_id)),
	                     NEW.chapter_num,
	                     (SELECT title FROM books WHERE id = NEW.book_id)),
	    deadline_epoch = NEW.deadline,
	    last_modified_epoch = strftime('%s','now'),
	    seq = seq + 1
	WHERE kind = 'chapter' AND source_id = NEW.id;
END;

DROP TRIGGER IF EXISTS cal_course_upd_name_code_chapters;
CREATE TRIGGER cal_course_upd_name_code_chapters
AFTER UPDATE OF name, code ON courses
BEGIN
	UPDATE calendar_index
	SET summary = printf('[%s] %s — Chapter %d — %s',
	                     NEW.code, NEW.name,
	                     (SELECT chapter_num FROM chapters WHERE id = calendar_index.source_id),
	                     (SELECT b.title FROM chapters ch JOIN books b ON b.id = ch.book_id WHERE ch.id = calendar_index.source_id)),
	    last_modified_epoch = strftime('%s','now'),
	    seq = seq + 1
	WHERE kind = 'chapter'
	  AND source_id IN (
	    SELECT ch.id
	    FROM chapters ch
	    JOIN books b ON b.id = ch.book_id
	    WHERE b.course_id = NEW.id
	  );
END;
`)
	if err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration tx: %w", err)
	}
	return nil
}
