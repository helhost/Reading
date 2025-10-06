package membership

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strings"

  "example.com/sqlite-server/session"
  "example.com/sqlite-server/util"
)

func RegisterMembershipRoutes(mux *http.ServeMux, db *sql.DB) {
  mux.HandleFunc("/user-universities", userUniversitiesHandler(db))
}

// Dispatcher for /user-universities
func userUniversitiesHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
      session.RequireAuth(db, getUserUniversitiesHandler(db)).ServeHTTP(w, r)
    case http.MethodPost:
      session.RequireAuth(db, postUserUniversityHandler(db)).ServeHTTP(w, r)
    default:
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
  }
}

// GET /user-universities  (list my memberships)
func getUserUniversitiesHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    userID, ok := session.UserIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }
    list, err := ListMemberships(db, userID)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    util.WriteJSON(w, list, http.StatusOK)
  }
}

// POST /user-universities  (subscribe me to a university)
func postUserUniversityHandler(db *sql.DB) http.HandlerFunc {
  type payload struct {
    UniversityID string `json:"universityId"`
  }
  return func(w http.ResponseWriter, r *http.Request) {
    userID, ok := session.UserIDFromCtx(r.Context())
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

    uniID := strings.TrimSpace(p.UniversityID)
    if uniID == "" {
      http.Error(w, "universityId is required", http.StatusBadRequest)
      return
    }

    created, m, err := AddMembership(db, userID, uniID)
    if err != nil {
      if err == sql.ErrNoRows {
        http.Error(w, "university not found", http.StatusBadRequest)
        return
      }
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    if created {
      util.WriteJSON(w, m, http.StatusCreated)
      return
    }
    util.WriteJSON(w, m, http.StatusOK) // idempotent re-subscribe
  }
}
