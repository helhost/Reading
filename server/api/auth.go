package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

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

func RegisterAuthRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/register", registerHandler(db))
	mux.HandleFunc("/login", loginHandler(db))
}

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

		w.WriteHeader(http.StatusCreated)
	}
}

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

		// success: respond with minimal user info (no session yet)
		writeJSON(w, loginResponse{UserID: u.ID}, http.StatusOK)
	}
}
