package store

import (
  "database/sql"
	"strings"
)
type Course struct {
	ID    int64   `json:"id"`
	Year  int64   `json:"year"`
	Term  int64   `json:"term"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Books []Item  `json:"books,omitempty"`
}


// GetCourseByID returns a single course by ID.
func GetCourseByID(db *sql.DB, id int64) (Course, error) {
	var c Course
	err := db.QueryRow(`
		SELECT id, year, term, code, name
		FROM courses
		WHERE id = ?;
	`, id).Scan(&c.ID, &c.Year, &c.Term, &c.Code, &c.Name)
	if err != nil {
		return Course{}, err
	}

	books, err := GetBooksByCourseID(db, c.ID)
	if err != nil {
		return Course{}, err
	}
	c.Books = books
	return c, nil
}

// GetAllCourses returns every course row.
func GetAllCourses(db *sql.DB) ([]Course, error) {
  // 1) load all courses
  rows, err := db.Query(`
    SELECT id, year, term, code, name
    FROM courses
    ORDER BY year DESC, term DESC, code;
  `)
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

  // 2) load all books for those course IDs in one go
  placeholders := make([]string, 0, len(ids))
  args := make([]any, 0, len(ids))
  for _, id := range ids {
    placeholders = append(placeholders, "?")
    args = append(args, id)
  }
  q := `
    SELECT
      b.id,
      b.title,
      b.author,
      COALESCE(b.numChapters, 0),
      COALESCE(b.completedChapters, 0),
      b.link,
      b.course_id
    FROM books b
    WHERE b.course_id IN (` + strings.Join(placeholders, ",") + `)
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
      &it.CompletedChapters,
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

  // 3) attach
  for i := range courses {
    courses[i].Books = byCourse[courses[i].ID]
  }
  return courses, nil
}


// AddCourse inserts a new course and returns it.
func AddCourse(db *sql.DB, year, term int64, code, name string) (Course, error) {
  res, err := db.Exec(`
    INSERT INTO courses (year, term, code, name)
    VALUES (?, ?, ?, ?);
  `, year, term, code, name)
  if err != nil {
    return Course{}, err
  }

  id, err := res.LastInsertId()
  if err != nil {
    return Course{}, err
  }

  var c Course
  err = db.QueryRow(`SELECT id, year, term, code, name FROM courses WHERE id = ?;`, id).
    Scan(&c.ID, &c.Year, &c.Term, &c.Code, &c.Name)
  if err != nil {
    return Course{}, err
  }
  return c, nil
}


// GetBooksByCourseID returns all books linked to a specific course.
// Item.Course is left nil to avoid recursive JSON structures.
func GetBooksByCourseID(db *sql.DB, courseID int64) ([]Item, error) {
	rows, err := db.Query(`
		SELECT
			b.id,
			b.title,
			b.author,
			COALESCE(b.numChapters, 0),
			COALESCE(b.completedChapters, 0),
			b.link
		FROM books b
		WHERE b.course_id = ?
		ORDER BY b.id;
	`, courseID)
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
			&it.CompletedChapters,
			&it.Link,
		); err != nil {
			return nil, err
		}
		// Avoid recursion: do not set it.Course here
		out = append(out, it)
	}
	return out, rows.Err()
}


func DeleteCourse(db *sql.DB, id int64) error {
  _, err := db.Exec(`DELETE FROM courses WHERE id = ?;`, id)
  return err
}
