package university

import "database/sql"

type University struct {
  ID        string `json:"id"`
  Name      string `json:"name"`
  CreatedAt int64  `json:"created_at"`
}

func AddUniversity(db *sql.DB, id, name string) (University, error) {
  _, err := db.Exec(`
    INSERT INTO universities (id, name)
    VALUES (?, ?)
  `, id, name)
  if err != nil {
    return University{}, err
  }

  var u University
  err = db.QueryRow(`
    SELECT id, name, created_at
      FROM universities
     WHERE id = ?
  `, id).Scan(&u.ID, &u.Name, &u.CreatedAt)
  if err != nil {
    return University{}, err
  }
  return u, nil
}
