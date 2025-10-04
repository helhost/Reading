package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"example.com/sqlite-server/store"
)

type ctxKey int

const ctxUserID ctxKey = iota

// readSessionID tries both cookie names (dev/prod).
func readSessionID(r *http.Request) string {
	if c, err := r.Cookie("__Host-session"); err == nil && c.Value != "" {
		return c.Value
	}
	if c, err := r.Cookie("session"); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}

func requireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := readSessionID(r)
		if sid == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		sess, err := store.GetSessionByID(db, sid)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		now := time.Now().Unix()
		if now >= sess.ExpiresAt {
			_ = store.DeleteSessionByID(db, sid) // cleanup if expired
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, sess.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func userIDFromCtx(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUserID)
	s, ok := v.(string)
	return s, ok && s != ""
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
		uid, ok := userIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		u, err := store.GetUserByID(db, uid)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		writeJSON(w, meResp{UserID: u.ID, Email: u.Email}, http.StatusOK)
	}
}
