package membership

import "database/sql"

type Membership struct {
  UserID       string `json:"userId"`
  UniversityID string `json:"universityId"`
  Role         string `json:"role"`
}

type MembershipView struct {
  UniversityID string `json:"universityId"`
  Name         string `json:"name"`
  Role         string `json:"role"`
  CreatedAt    int64  `json:"created_at"`
}

// AddMembership subscribes user to a university (idempotent).
// Returns (created, Membership, error).
func AddMembership(db *sql.DB, userID, universityID string) (bool, Membership, error) {
  // Ensure the university exists
  var exists string
  if err := db.QueryRow(`SELECT id FROM universities WHERE id = ?`, universityID).Scan(&exists); err != nil {
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

  // Read final role
  var role string
  if err := db.QueryRow(`
    SELECT role
      FROM user_universities
     WHERE user_id = ? AND university_id = ?
  `, userID, universityID).Scan(&role); err != nil {
    return created, Membership{}, err
  }

  return created, Membership{UserID: userID, UniversityID: universityID, Role: role}, nil
}

// RemoveMembership unsubscribes user from a university.
// Returns (deleted, error). Idempotent: deleted=false if nothing to remove.
func RemoveMembership(db *sql.DB, userID, universityID string) (bool, error) {
  res, err := db.Exec(`
    DELETE FROM user_universities
     WHERE user_id = ? AND university_id = ?
  `, userID, universityID)
  if err != nil {
    return false, err
  }
  n, _ := res.RowsAffected()
  return n > 0, nil
}

// ListMemberships returns all universities the user is a member of.
func ListMemberships(db *sql.DB, userID string) ([]MembershipView, error) {
  rows, err := db.Query(`
    SELECT u.id, u.name, uu.role, u.created_at
      FROM universities u
      JOIN user_universities uu
        ON uu.university_id = u.id
     WHERE uu.user_id = ?
     ORDER BY u.name ASC
  `, userID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  out := make([]MembershipView, 0, 16)
  for rows.Next() {
    var mv MembershipView
    if err := rows.Scan(&mv.UniversityID, &mv.Name, &mv.Role, &mv.CreatedAt); err != nil {
      return nil, err
    }
    out = append(out, mv)
  }
  return out, rows.Err()
}


// IsMember reports whether userID is a member of universityID.
func IsMember(db *sql.DB, userID, universityID string) (bool, error) {
	var x int
	err := db.QueryRow(`
		SELECT 1
		  FROM user_universities
		 WHERE user_id = ? AND university_id = ?
		 LIMIT 1
	`, userID, universityID).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
