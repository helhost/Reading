package membership

import (
  "database/sql"
)

type Membership struct {
  UserID       string `json:"userId"`
  UniversityID string `json:"universityId"`
  Role         string `json:"role"`
}

// AddMembership subscribes user to university (idempotent).
// Returns (created, Membership, error).
func AddMembership(db *sql.DB, userID, universityID string) (bool, Membership, error) {
  // Ensure university exists
  var tmp string
  if err := db.QueryRow(`SELECT id FROM universities WHERE id = ?`, universityID).Scan(&tmp); err != nil {
    if err == sql.ErrNoRows {
      return false, Membership{}, sql.ErrNoRows
    }
    return false, Membership{}, err
  }

  // Insert or ignore to be idempotent
  res, err := db.Exec(`
    INSERT OR IGNORE INTO user_universities (user_id, university_id, role)
    VALUES (?, ?, 'member')
  `, userID, universityID)
  if err != nil {
    return false, Membership{}, err
  }

  created := false
  if n, _ := res.RowsAffected(); n > 0 {
    created = true
  }

  // Read back role (in case defaults/changes later)
  var role string
  if err := db.QueryRow(`
    SELECT role FROM user_universities
    WHERE user_id = ? AND university_id = ?
  `, userID, universityID).Scan(&role); err != nil {
    return created, Membership{}, err
  }

  return created, Membership{UserID: userID, UniversityID: universityID, Role: role}, nil
}
