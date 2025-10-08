package session

import (
  "context"
  "database/sql"
  "net/http"
  "time"
)

// private context key type
type ctxKey int

const ctxUserID ctxKey = iota

// ReadSessionID is exported so other packages can reuse it if needed.
func ReadSessionID(r *http.Request) string {
  if c, err := r.Cookie("__Host-session"); err == nil && c.Value != "" {
    return c.Value
  }
  if c, err := r.Cookie("session"); err == nil && c.Value != "" {
    return c.Value
  }
  return ""
}

// RequireAuth is middleware that validates the session cookie,
// loads the session, and injects userID into the request context.
func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    sid := ReadSessionID(r)
    if sid == "" {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    // Use same-package functions from sessionService.go
    sess, err := GetSessionByID(db, sid)
    if err != nil {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    now := time.Now().Unix()
    if now >= sess.ExpiresAt {
      _ = DeleteSessionByID(db, sid) // cleanup if expired
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    ctx := context.WithValue(r.Context(), ctxUserID, sess.UserID)
    next.ServeHTTP(w, r.WithContext(ctx))
  }
}

// UserIDFromCtx extracts the userID set by RequireAuth.
func UserIDFromCtx(ctx context.Context) (string, bool) {
  v := ctx.Value(ctxUserID)
  s, ok := v.(string)
  return s, ok && s != ""
}
