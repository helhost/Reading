package store

import (
  "database/sql"

  _ "modernc.org/sqlite"
)

func OpenDB(dsn string) (*sql.DB, error) {
  db, err := sql.Open("sqlite", dsn)
  if err != nil {
    return nil, err
  }
  // tiny app: keep connections low
  db.SetMaxOpenConns(1) // only 1 server should be active at a time
  db.SetMaxIdleConns(1)
  return db, nil
}

func EnsureSchema(db *sql.DB) error {
  _, err := db.Exec(`
    PRAGMA foreign_keys = ON;

    -- Core identities
    CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      email TEXT NOT NULL UNIQUE,
      password TEXT NOT NULL,
      created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
    );

		CREATE TABLE IF NOT EXISTS admins (
			user_id TEXT PRIMARY KEY,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

    CREATE TABLE IF NOT EXISTS universities (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL UNIQUE,
      created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
    );

    CREATE TABLE IF NOT EXISTS user_universities (
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      university_id TEXT NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
      role TEXT NOT NULL DEFAULT 'member',
      PRIMARY KEY (user_id, university_id)
    );

    -- Courses and enrolment
    CREATE TABLE IF NOT EXISTS courses (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      university_id TEXT NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
      year INTEGER NOT NULL,
      term INTEGER NOT NULL,
      code TEXT NOT NULL,
      name TEXT NOT NULL,
      UNIQUE (university_id, year, term, code)
    );

    CREATE TABLE IF NOT EXISTS user_courses (
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
      PRIMARY KEY (user_id, course_id)
    );

    -- Materials
    CREATE TABLE IF NOT EXISTS books (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      course_id INTEGER REFERENCES courses(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      author TEXT NOT NULL,
      numChapters INTEGER,
      location TEXT
    );

    CREATE TABLE IF NOT EXISTS chapters (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
      chapter_num INTEGER NOT NULL,
      deadline INTEGER
    );

    CREATE TABLE IF NOT EXISTS articles (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      course_id INTEGER REFERENCES courses(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      author TEXT NOT NULL,
      location TEXT,
      deadline INTEGER
    );

    CREATE TABLE IF NOT EXISTS assignments (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      course_id INTEGER REFERENCES courses(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      description TEXT,
      deadline INTEGER
    );

    -- Unified progress (exactly one FK set)
    CREATE TABLE IF NOT EXISTS progress (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      chapter_id INTEGER REFERENCES chapters(id) ON DELETE CASCADE,
      article_id INTEGER REFERENCES articles(id) ON DELETE CASCADE,
      assignment_id INTEGER REFERENCES assignments(id) ON DELETE CASCADE,
      completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0,1)),
      CHECK (
        (chapter_id IS NOT NULL AND article_id IS NULL AND assignment_id IS NULL) OR
        (chapter_id IS NULL AND article_id IS NOT NULL AND assignment_id IS NULL) OR
        (chapter_id IS NULL AND article_id IS NULL AND assignment_id IS NOT NULL)
      ),
      UNIQUE (user_id, chapter_id),
      UNIQUE (user_id, article_id),
      UNIQUE (user_id, assignment_id)
    );

    -- Sessions last (depends on users)
    CREATE TABLE IF NOT EXISTS sessions (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
      expires_at INTEGER NOT NULL
    );

    -- Indexes
    CREATE INDEX IF NOT EXISTS idx_user_universities_university
      ON user_universities(university_id);

    CREATE INDEX IF NOT EXISTS idx_courses_university
      ON courses(university_id);

    CREATE INDEX IF NOT EXISTS idx_courses_university_year_term
      ON courses(university_id, year, term);

    CREATE INDEX IF NOT EXISTS idx_user_courses_course
      ON user_courses(course_id);

    CREATE INDEX IF NOT EXISTS idx_books_course
      ON books(course_id);

    CREATE INDEX IF NOT EXISTS idx_chapters_book
      ON chapters(book_id);

    CREATE INDEX IF NOT EXISTS idx_chapters_deadline
      ON chapters(deadline);

    CREATE INDEX IF NOT EXISTS idx_articles_course
      ON articles(course_id);

    CREATE INDEX IF NOT EXISTS idx_articles_deadline
      ON articles(deadline);

    CREATE INDEX IF NOT EXISTS idx_assignments_course
      ON assignments(course_id);

    CREATE INDEX IF NOT EXISTS idx_assignments_deadline
      ON assignments(deadline);

    CREATE INDEX IF NOT EXISTS idx_progress_user
      ON progress(user_id);

    CREATE INDEX IF NOT EXISTS idx_progress_chapter
      ON progress(chapter_id)
      WHERE chapter_id IS NOT NULL;

    CREATE INDEX IF NOT EXISTS idx_progress_article
      ON progress(article_id)
      WHERE article_id IS NOT NULL;

    CREATE INDEX IF NOT EXISTS idx_progress_assignment
      ON progress(assignment_id)
      WHERE assignment_id IS NOT NULL;

    CREATE INDEX IF NOT EXISTS idx_sessions_user
      ON sessions(user_id);

    CREATE INDEX IF NOT EXISTS idx_sessions_expires
      ON sessions(expires_at);
  `)
  return err
}
