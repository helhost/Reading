package auth

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strings"
  "time"

  "github.com/google/uuid"
  "golang.org/x/crypto/bcrypt"

  "example.com/sqlite-server/session"
  "example.com/sqlite-server/util"
)

type registerPayload struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

type loginPayload struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

type loginResponse struct {
  UserID string `json:"userId"`
}

// RegisterAuthRoutes wires up the auth endpoints.
func RegisterAuthRoutes(mux *http.ServeMux, db *sql.DB) {
  mux.HandleFunc("/register", registerHandler(db))
  mux.HandleFunc("/login", loginHandler(db))
  mux.HandleFunc("/logout", logoutHandler(db))
  mux.HandleFunc("/me", session.RequireAuth(db, meHandler(db))) // use exported middleware
}

// POST /register
func registerHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }

    var p registerPayload
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&p); err != nil {
      http.Error(w, "bad request", http.StatusBadRequest)
      return
    }

    email := strings.ToLower(strings.TrimSpace(p.Email))
    if !util.VerifyEmail(email) {
      http.Error(w, "invalid email", http.StatusBadRequest)
      return
    }
    if len(p.Password) < 8 {
      http.Error(w, "password too short (min 8)", http.StatusBadRequest)
      return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    id := uuid.NewString()
    if err := AddUser(db, id, email, string(hash)); err != nil { // same package
      if strings.Contains(strings.ToLower(err.Error()), "unique") {
        http.Error(w, "email already registered", http.StatusConflict)
        return
      }
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    sess, err := session.CreateSession(db, id, 7*24*time.Hour)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    secure := util.IsProd()
    name := "session"
    if secure { name = "__Host-session" }

    http.SetCookie(w, &http.Cookie{
      Name:     name,
      Value:    sess.ID,
      Path:     "/",
      HttpOnly: true,
      SameSite: http.SameSiteStrictMode,
      Secure:   secure,
      Expires:  time.Unix(sess.ExpiresAt, 0),
    })

    util.WriteJSON(w, loginResponse{UserID: id}, http.StatusCreated)
  }
}

// POST /login
func loginHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }

    var p loginPayload
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&p); err != nil {
      http.Error(w, "bad request", http.StatusBadRequest)
      return
    }

    email := strings.ToLower(strings.TrimSpace(p.Email))
    if !util.VerifyEmail(email) || len(p.Password) == 0 {
      http.Error(w, "invalid credentials", http.StatusUnauthorized)
      return
    }

    u, err := GetUserByEmail(db, email) // same package
    if err != nil {
      http.Error(w, "invalid credentials", http.StatusUnauthorized)
      return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p.Password)); err != nil {
      http.Error(w, "invalid credentials", http.StatusUnauthorized)
      return
    }

    sess, err := session.CreateSession(db, u.ID, 7*24*time.Hour)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    secure := util.IsProd()
    name := "session"
    if secure { name = "__Host-session" }

    http.SetCookie(w, &http.Cookie{
      Name:     name,
      Value:    sess.ID,
      Path:     "/",
      HttpOnly: true,
      SameSite: http.SameSiteStrictMode,
      Secure:   secure,
      Expires:  time.Unix(sess.ExpiresAt, 0),
    })

    util.WriteJSON(w, loginResponse{UserID: u.ID}, http.StatusOK)
  }
}

// POST /logout
func logoutHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }

    // Best-effort: read whichever cookie exists and delete that session.
    var sid string
    if c, err := r.Cookie("__Host-session"); err == nil && c.Value != "" {
      sid = c.Value
    } else if c, err := r.Cookie("session"); err == nil && c.Value != "" {
      sid = c.Value
    }
    if sid != "" {
      _ = session.DeleteSessionByID(db, sid)
    }

    // Clear both possible cookie names (dev/prod).
    for _, name := range []string{"session", "__Host-session"} {
      http.SetCookie(w, &http.Cookie{
        Name:     name,
        Value:    "",
        Path:     "/",
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
        HttpOnly: true,
        SameSite: http.SameSiteStrictMode,
        Secure:   util.IsProd(),
      })
    }

    w.WriteHeader(http.StatusNoContent)
  }
}


// GET /me
// Returns { userId, email } for the current session.
func meHandler(db *sql.DB) http.HandlerFunc {
  type meResp struct {
    UserID string `json:"userId"`
    Email  string `json:"email"`
  }
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    uid, ok := session.UserIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }
    u, err := GetUserByID(db, uid)
    if err != nil {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }
    util.WriteJSON(w, meResp{UserID: u.ID, Email: u.Email}, http.StatusOK)
  }
}
