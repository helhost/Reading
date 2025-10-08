package university

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strings"

  "github.com/google/uuid"

  "example.com/sqlite-server/session"
  "example.com/sqlite-server/util"
)

func RegisterUniversityRoutes(mux *http.ServeMux, db *sql.DB) {
  mux.HandleFunc("/universities", universitiesHandler(db))
}

// Dispatcher — GET is public; POST requires auth
func universitiesHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
      getUniversitiesHandler(db)(w, r)
    case http.MethodPost:
      session.RequireAuth(db, postUniversityHandler(db)).ServeHTTP(w, r)
    case http.MethodDelete:
      session.RequireAuth(db, deleteUniversityHandler(db)).ServeHTTP(w, r)
    default:
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
  }
}

// GET /universities (public)
func getUniversitiesHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    list, err := ListUniversities(db)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    util.WriteJSON(w, list, http.StatusOK)
  }
}

// POST /universities (auth required)
func postUniversityHandler(db *sql.DB) http.HandlerFunc {
  type payload struct {
    Name string `json:"name"`
  }
  return func(w http.ResponseWriter, r *http.Request) {
    var p payload
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&p); err != nil {
      http.Error(w, "bad request", http.StatusBadRequest)
      return
    }

    name := strings.TrimSpace(p.Name)
    if name == "" {
      http.Error(w, "name is required", http.StatusBadRequest)
      return
    }

    id := uuid.NewString()
    uni, err := AddUniversity(db, id, name)
    if err != nil {
      if strings.Contains(strings.ToLower(err.Error()), "unique") {
        http.Error(w, "university name already exists", http.StatusConflict)
        return
      }
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    util.WriteJSON(w, uni, http.StatusCreated)
  }
}


// DELETE /universities
// Body: { "universityId": "uuid" }
// Success: 204 No Content
// Errors: 400 invalid, 404 not found, 409 has courses, 500 internal
func deleteUniversityHandler(db *sql.DB) http.HandlerFunc {
  type payload struct {
    UniversityID string `json:"universityId"`
  }
  return func(w http.ResponseWriter, r *http.Request) {
    var p payload
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&p); err != nil {
      http.Error(w, "bad request", http.StatusBadRequest)
      return
    }
    id := strings.TrimSpace(p.UniversityID)
    if id == "" {
      http.Error(w, "universityId is required", http.StatusBadRequest)
      return
    }

    deleted, err := DeleteUniversityIfNoCourses(db, id)
    if err != nil {
      lc := strings.ToLower(err.Error())
      switch {
      case err == sql.ErrNoRows:
        http.Error(w, "not found", http.StatusNotFound)
        return
      case strings.Contains(lc, "invalid input"):
        http.Error(w, "invalid input", http.StatusBadRequest)
        return
      case strings.Contains(lc, "has courses"):
        http.Error(w, "conflict: university has courses", http.StatusConflict)
        return
      default:
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
      }
    }
    if !deleted {
      // Shouldn’t normally happen (we already handle known errors), but be explicit.
      http.Error(w, "not found", http.StatusNotFound)
      return
    }
    w.WriteHeader(http.StatusNoContent)
  }
}
