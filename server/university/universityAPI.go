package university

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strings"

  "github.com/google/uuid"

  "example.com/sqlite-server/util"
  "example.com/sqlite-server/session"
)

func RegisterUniversityRoutes(mux *http.ServeMux, db *sql.DB) {
  mux.HandleFunc("/universities", session.RequireAuth(db, postUniversity(db)))
}

func postUniversity(db *sql.DB) http.HandlerFunc {
  type payload struct {
    Name string `json:"name"`
  }
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    if _, ok := session.UserIDFromCtx(r.Context()); !ok {
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
