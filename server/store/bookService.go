package store

import (
  "database/sql"
)

type Item struct {
  ID          int64   `json:"id"`
  Title       string  `json:"title"`
  Author      string  `json:"author"`
  NumChapters int64   `json:"numChapters"`
  Link        *string `json:"link,omitempty"`
  Course      *Course `json:"course,omitempty"`
}

// GetAllItems returns all books for a user, with optional course embedded.
func GetAllItems(db *sql.DB, userID string) ([]Item, error) {
  rows, err := db.Query(`
    SELECT
      b.id,
      b.title,
      b.author,
      COALESCE(b.numChapters, 0),
      b.link,
      c.id,
      c.year,
      c.term,
      c.code,
      c.name
    FROM books b
    LEFT JOIN courses c
      ON c.id = b.course_id
     AND c.user_id = b.user_id
    WHERE b.user_id = ?
    ORDER BY b.id;
  `, userID)
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

// GetBookByID returns a single book by id for a user.
func GetBookByID(db *sql.DB, userID string, id int64) (Item, error) {
  row := db.QueryRow(`
    SELECT
      b.id,
      b.title,
      b.author,
      COALESCE(b.numChapters, 0),
      b.link,
      c.id,
      c.year,
      c.term,
      c.code,
      c.name
    FROM books b
    LEFT JOIN courses c
      ON c.id = b.course_id
     AND c.user_id = b.user_id
    WHERE b.id = ? AND b.user_id = ?;
  `, id, userID)

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

// AddBook inserts a new book (optionally linked to a course) for a user.
func AddBook(db *sql.DB, userID, title, author string, numChapters int64, link *string, courseID *int64) (Item, error) {
  // If a course is provided, ensure it belongs to the same user.
  if courseID != nil {
    var ok int
    if err := db.QueryRow(`SELECT 1 FROM courses WHERE id = ? AND user_id = ?`, *courseID, userID).Scan(&ok); err != nil {
      if err == sql.ErrNoRows {
        return Item{}, sql.ErrNoRows // caller can map to 404/400
      }
      return Item{}, err
    }
  }

  _, err := db.Exec(`
    INSERT INTO books (user_id, title, author, numChapters, link, course_id)
    VALUES (?, ?, ?, ?, ?, ?);
  `, userID, title, author, numChapters, link, courseID)
  if err != nil {
    return Item{}, err
  }

  var id int64
  if err := db.QueryRow(`SELECT last_insert_rowid();`).Scan(&id); err != nil {
    return Item{}, err
  }
  return GetBookByID(db, userID, id)
}

func DeleteBook(db *sql.DB, userID string, id int64) error {
  res, err := db.Exec(`DELETE FROM books WHERE id = ? AND user_id = ?;`, id, userID)
  if err != nil {
    return err
  }
  if n, _ := res.RowsAffected(); n == 0 {
    return sql.ErrNoRows
  }
  return nil
}
