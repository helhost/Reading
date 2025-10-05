package store

import (
  "database/sql"
  "strings"
)

type Course struct {
  ID    int64  `json:"id"`
  Year  int64  `json:"year"`
  Term  int64  `json:"term"`
  Code  string `json:"code"`
  Name  string `json:"name"`
  Books []Item `json:"books,omitempty"`
}

// GetCourseByID returns a single course by ID for a user.
func GetCourseByID(db *sql.DB, userID string, id int64) (Course, error) {
  var c Course
  err := db.QueryRow(`
    SELECT id, year, term, code, name
      FROM courses
     WHERE id = ? AND user_id = ?;
  `, id, userID).Scan(&c.ID, &c.Year, &c.Term, &c.Code, &c.Name)
  if err != nil {
    return Course{}, err
  }

  books, err := GetBooksByCourseID(db, userID, c.ID)
  if err != nil {
    return Course{}, err
  }
  c.Books = books
  return c, nil
}

// GetAllCourses returns all courses for a user (with their books).
func GetAllCourses(db *sql.DB, userID string) ([]Course, error) {
  rows, err := db.Query(`
    SELECT id, year, term, code, name
      FROM courses
     WHERE user_id = ?
     ORDER BY year DESC, term DESC, code;
  `, userID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  courses := make([]Course, 0, 64)
  ids := make([]int64, 0, 64)
  for rows.Next() {
    var c Course
    if err := rows.Scan(&c.ID, &c.Year, &c.Term, &c.Code, &c.Name); err != nil {
      return nil, err
    }
    ids = append(ids, c.ID)
    courses = append(courses, c)
  }
  if err := rows.Err(); err != nil {
    return nil, err
  }
  if len(courses) == 0 {
    return courses, nil
  }

  // load all books for those course IDs in one go (scoped by user)
  placeholders := make([]string, 0, len(ids))
  args := make([]any, 0, len(ids)+1)
  args = append(args, userID)
  for range ids {
    placeholders = append(placeholders, "?")
  }
  for _, id := range ids {
    args = append(args, id)
  }

  q := `
    SELECT
      b.id,
      b.title,
      b.author,
      COALESCE(b.numChapters, 0),
      b.link,
      b.course_id
      FROM books b
     WHERE b.user_id = ?
       AND b.course_id IN (` + strings.Join(placeholders, ",") + `)
     ORDER BY b.course_id, b.id;
  `
  brows, err := db.Query(q, args...)
  if err != nil {
    return nil, err
  }
  defer brows.Close()

  byCourse := make(map[int64][]Item, len(ids))
  for brows.Next() {
    var it Item
    var courseID int64
    if err := brows.Scan(
      &it.ID,
      &it.Title,
      &it.Author,
      &it.NumChapters,
      &it.Link,
      &courseID,
    ); err != nil {
      return nil, err
    }
    it.Course = nil
    byCourse[courseID] = append(byCourse[courseID], it)
  }
  if err := brows.Err(); err != nil {
    return nil, err
  }

  for i := range courses {
    courses[i].Books = byCourse[courses[i].ID]
  }
  return courses, nil
}

// AddCourse inserts a new course for a user and returns it.
func AddCourse(db *sql.DB, userID string, year, term int64, code, name string) (Course, error) {
  res, err := db.Exec(`
    INSERT INTO courses (user_id, year, term, code, name)
    VALUES (?, ?, ?, ?, ?);
  `, userID, year, term, code, name)
  if err != nil {
    return Course{}, err
  }

  id, err := res.LastInsertId()
  if err != nil {
    return Course{}, err
  }

  var c Course
  err = db.QueryRow(`
    SELECT id, year, term, code, name
      FROM courses
     WHERE id = ? AND user_id = ?;
  `, id, userID).Scan(&c.ID, &c.Year, &c.Term, &c.Code, &c.Name)
  if err != nil {
    return Course{}, err
  }
  return c, nil
}

// GetBooksByCourseID returns all books for a specific course owned by user.
func GetBooksByCourseID(db *sql.DB, userID string, courseID int64) ([]Item, error) {
  rows, err := db.Query(`
    SELECT
      b.id,
      b.title,
      b.author,
      COALESCE(b.numChapters, 0),
      b.link
      FROM books b
     WHERE b.user_id = ?
       AND b.course_id = ?
     ORDER BY b.id;
  `, userID, courseID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var out []Item
  for rows.Next() {
    var it Item
    if err := rows.Scan(
      &it.ID,
      &it.Title,
      &it.Author,
      &it.NumChapters,
      &it.Link,
    ); err != nil {
      return nil, err
    }
    out = append(out, it)
  }
  return out, rows.Err()
}

func DeleteCourse(db *sql.DB, userID string, id int64) error {
  res, err := db.Exec(`DELETE FROM courses WHERE id = ? AND user_id = ?;`, id, userID)
  if err != nil {
    return err
  }
  if n, _ := res.RowsAffected(); n == 0 {
    return sql.ErrNoRows
  }
  return nil
}
