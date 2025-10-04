package store

import (
  "database/sql"
)

type Item struct {
  ID                int64    `json:"id"`
  Title             string   `json:"title"`
  Author            string   `json:"author"`
  NumChapters       int64    `json:"numChapters"`
  CompletedChapters int64    `json:"completedChapters"`
  Link              *string  `json:"link,omitempty"`
  Course            *Course  `json:"course,omitempty"`
}

// GetAllItems returns every row, with optional course embedded.
func GetAllItems(db *sql.DB) ([]Item, error) {
  rows, err := db.Query(`
    SELECT
      b.id,
      b.title,
      b.author,
      COALESCE(b.numChapters, 0),
      COALESCE(b.completedChapters, 0),
      b.link,
      c.id,
      c.year,
      c.term,
      c.code,
      c.name
    FROM books b
    LEFT JOIN courses c ON c.id = b.course_id
    ORDER BY b.id;
  `)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var out []Item
  for rows.Next() {
    var it Item
    var (
      cID   sql.NullInt64
      cYear sql.NullInt64
      cTerm sql.NullInt64
      cCode sql.NullString
      cName sql.NullString
    )
    if err := rows.Scan(
      &it.ID,
      &it.Title,
      &it.Author,
      &it.NumChapters,
      &it.CompletedChapters,
      &it.Link,
      &cID, &cYear, &cTerm, &cCode, &cName,
    ); err != nil {
      return nil, err
    }
    if cID.Valid {
      it.Course = &Course{
        ID:   cID.Int64,
        Year: cYear.Int64,
        Term: cTerm.Int64,
        Code: cCode.String,
        Name: cName.String,
      }
    }
    out = append(out, it)
  }
  return out, rows.Err()
}

// GetBookByID returns a single book with embedded course if present.
func GetBookByID(db *sql.DB, id int64) (Item, error) {
  row := db.QueryRow(`
    SELECT
      b.id,
      b.title,
      b.author,
      COALESCE(b.numChapters, 0),
      COALESCE(b.completedChapters, 0),
      b.link,
      c.id,
      c.year,
      c.term,
      c.code,
      c.name
    FROM books b
    LEFT JOIN courses c ON c.id = b.course_id
    WHERE b.id = ?;
  `, id)

  var it Item
  var (
    cID   sql.NullInt64
    cYear sql.NullInt64
    cTerm sql.NullInt64
    cCode sql.NullString
    cName sql.NullString
  )
  if err := row.Scan(
    &it.ID,
    &it.Title,
    &it.Author,
    &it.NumChapters,
    &it.CompletedChapters,
    &it.Link,
    &cID, &cYear, &cTerm, &cCode, &cName,
  ); err != nil {
    return Item{}, err
  }
  if cID.Valid {
    it.Course = &Course{
      ID:   cID.Int64,
      Year: cYear.Int64,
      Term: cTerm.Int64,
      Code: cCode.String,
      Name: cName.String,
    }
  }
  return it, nil
}

// AddBook inserts a new book (optionally linked to a course) and returns it.
func AddBook(db *sql.DB, title, author string, numChapters int64, link *string, courseID *int64) (Item, error) {
  _, err := db.Exec(`
    INSERT INTO books (title, author, numChapters, completedChapters, link, course_id)
    VALUES (?, ?, ?, 0, ?, ?);
  `, title, author, numChapters, link, courseID)
  if err != nil {
    return Item{}, err
  }

  // Return the last inserted row using SQLite's last_insert_rowid()
  var id int64
  if err := db.QueryRow(`SELECT last_insert_rowid();`).Scan(&id); err != nil {
    return Item{}, err
  }
  return GetBookByID(db, id)
}

// UpdateCompletedChapters sets completedChapters for a book and returns the updated row.
func UpdateCompletedChapters(db *sql.DB, id int64, completed int64) (Item, error) {
  if _, err := db.Exec(`
    UPDATE books
    SET completedChapters = ?
    WHERE id = ?;
  `, completed, id); err != nil {
    return Item{}, err
  }
  return GetBookByID(db, id)
}

func DeleteBook(db *sql.DB, id int64) error {
  _, err := db.Exec(`DELETE FROM books WHERE id = ?;`, id)
  return err
}
