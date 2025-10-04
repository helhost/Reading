package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"example.com/sqlite-server/store"
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
	mux.HandleFunc("/me", requireAuth(db, meHandler(db)))
}

// POST /register
// Body: { "email": "...", "password": "..." }
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
		if !verifyEmail(email) {
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
		if err := store.AddUser(db, id, email, string(hash)); err != nil {
			// crude UNIQUE(email) detection for SQLite
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				http.Error(w, "email already registered", http.StatusConflict)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		sess, err := store.CreateSession(db, id, 7*24*time.Hour)
		if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
		}

		secure := isProd()
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

		writeJSON(w, loginResponse{UserID: id}, http.StatusCreated)
	}
}

// POST /login
// Body: { "email": "...", "password": "..." }
// Sets the session cookie and returns { userId }
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
		if !verifyEmail(email) || len(p.Password) == 0 {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		u, err := store.GetUserByEmail(db, email)
		if err != nil {
			// avoid user enumeration: same message for not found
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p.Password)); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create a new session (7-day TTL).
		sess, err := store.CreateSession(db, u.ID, 7*24*time.Hour)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		secure := isProd()
		name := "session"
		if secure {
			name = "__Host-session"
		}

		http.SetCookie(w, &http.Cookie{
			Name:     name,                          // "__Host-session" in prod
			Value:    sess.ID,                      // opaque token
			Path:     "/",                          // send to all routes
			HttpOnly: true,                         // JS can't read
			SameSite: http.SameSiteStrictMode,      // mitigates CSRF on same-origin
			Secure:   secure,                       // true in prod (HTTPS)
			Expires:  time.Unix(sess.ExpiresAt, 0), // aligns with DB
		})

		writeJSON(w, loginResponse{UserID: u.ID}, http.StatusOK)
	}
}

// POST /logout
// Clears the cookie and deletes the server-side session if present.
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
			_ = store.DeleteSessionByID(db, sid)
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
				Secure:   isProd(),
			})
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
