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
