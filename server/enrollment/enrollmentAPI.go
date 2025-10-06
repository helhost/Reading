package enrollment

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strconv"
  "strings"

  "example.com/sqlite-server/membership"
  "example.com/sqlite-server/session"
  "example.com/sqlite-server/util"
)

func RegisterEnrollmentRoutes(mux *http.ServeMux, db *sql.DB) {
  // Auth required; dispatcher handles methods
  mux.HandleFunc("/user-courses", session.RequireAuth(db, userCoursesHandler(db)))
}

func userCoursesHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
      postUserCourseHandler(db)(w, r)
    case http.MethodDelete:
      deleteUserCourseHandler(db)(w, r)
    default:
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
  }
}

// POST /user-courses
// Body: { "courseId": 123 }  (accepts "123" or 123)
func postUserCourseHandler(db *sql.DB) http.HandlerFunc {
  type payload struct {
    CourseID any `json:"courseId"`
  }
  return func(w http.ResponseWriter, r *http.Request) {
    uid, ok := session.UserIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    var p payload
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&p); err != nil {
      http.Error(w, "bad request", http.StatusBadRequest)
      return
    }

    var cid int64
    switch v := p.CourseID.(type) {
    case float64:
      cid = int64(v)
    case string:
      s := strings.TrimSpace(v)
      if s == "" {
        http.Error(w, "courseId is required", http.StatusBadRequest)
        return
      }
      n, err := strconv.ParseInt(s, 10, 64)
      if err != nil {
        http.Error(w, "invalid courseId", http.StatusBadRequest)
        return
      }
      cid = n
    default:
      http.Error(w, "invalid courseId", http.StatusBadRequest)
      return
    }
    if cid <= 0 {
      http.Error(w, "invalid courseId", http.StatusBadRequest)
      return
    }

    uniID, err := CourseUniversity(db, cid)
    if err != nil {
      if err == sql.ErrNoRows {
        http.Error(w, "course not found", http.StatusBadRequest)
        return
      }
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    isMember, err := membership.IsMember(db, uid, uniID)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    if !isMember {
      http.Error(w, "forbidden", http.StatusForbidden)
      return
    }

    created, e, err := AddEnrollment(db, uid, cid)
    if err != nil {
      if err == sql.ErrNoRows {
        http.Error(w, "course not found", http.StatusBadRequest)
        return
      }
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    if created {
      util.WriteJSON(w, e, http.StatusCreated)
      return
    }
    util.WriteJSON(w, e, http.StatusOK) // idempotent
  }
}

// DELETE /user-courses
// Body: { "courseId": 123 }  (accepts "123" or 123)
// Idempotent: always 204 if input is valid (even if no enrollment existed).
func deleteUserCourseHandler(db *sql.DB) http.HandlerFunc {
  type payload struct {
    CourseID any `json:"courseId"`
  }
  return func(w http.ResponseWriter, r *http.Request) {
    uid, ok := session.UserIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    var p payload
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&p); err != nil {
      http.Error(w, "bad request", http.StatusBadRequest)
      return
    }

    var cid int64
    switch v := p.CourseID.(type) {
    case float64:
      cid = int64(v)
    case string:
      s := strings.TrimSpace(v)
      if s == "" {
        http.Error(w, "courseId is required", http.StatusBadRequest)
        return
      }
      n, err := strconv.ParseInt(s, 10, 64)
      if err != nil {
        http.Error(w, "invalid courseId", http.StatusBadRequest)
        return
      }
      cid = n
    default:
      http.Error(w, "invalid courseId", http.StatusBadRequest)
      return
    }
    if cid <= 0 {
      http.Error(w, "invalid courseId", http.StatusBadRequest)
      return
    }

    // Confirm course exists and user is a member of its university.
    uniID, err := CourseUniversity(db, cid)
    if err != nil {
      if err == sql.ErrNoRows {
        // Treat nonexistent course as a client error (same as POST)
        http.Error(w, "course not found", http.StatusBadRequest)
        return
      }
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    isMember, err := membership.IsMember(db, uid, uniID)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    if !isMember {
      http.Error(w, "forbidden", http.StatusForbidden)
      return
    }

    // Idempotent remove.
    _, err = RemoveEnrollment(db, uid, cid)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    w.WriteHeader(http.StatusNoContent)
  }
}
