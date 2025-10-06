
package enrollment

import "database/sql"

type Enrollment struct {
  UserID   string `json:"userId"`
  CourseID int64  `json:"courseId"`
}

// AddEnrollment subscribes a user to a course (idempotent).
// It assumes the caller has already passed authorization checks.
func AddEnrollment(db *sql.DB, userID string, courseID int64) (bool, Enrollment, error) {
  // Ensure course exists (fail fast with 404/400 at the API layer)
  var exists int64
  if err := db.QueryRow(`SELECT id FROM courses WHERE id = ?`, courseID).Scan(&exists); err != nil {
    if err == sql.ErrNoRows {
      return false, Enrollment{}, sql.ErrNoRows
    }
    return false, Enrollment{}, err
  }

  // Idempotent insert
  res, err := db.Exec(`
    INSERT OR IGNORE INTO user_courses (user_id, course_id)
    VALUES (?, ?)
  `, userID, courseID)
  if err != nil {
    return false, Enrollment{}, err
  }
  created := false
  if n, _ := res.RowsAffected(); n > 0 {
    created = true
  }

  return created, Enrollment{UserID: userID, CourseID: courseID}, nil
}

func RemoveEnrollment(db *sql.DB, userID string, courseID int64) (bool, error) {
  res, err := db.Exec(`
    DELETE FROM user_courses
     WHERE user_id = ? AND course_id = ?
  `, userID, courseID)
  if err != nil {
    return false, err
  }
  n, _ := res.RowsAffected()
  return n > 0, nil
}

// CourseUniversity returns the university_id owning this course (for auth checks).
func CourseUniversity(db *sql.DB, courseID int64) (string, error) {
  var uniID string
  err := db.QueryRow(`SELECT university_id FROM courses WHERE id = ?`, courseID).Scan(&uniID)
  return uniID, err
}
