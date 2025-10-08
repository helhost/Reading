package calendar

import (
	"database/sql"
	"net/http"
	"strings"
	"strconv"
	"time"

	"example.com/sqlite-server/session"
	"example.com/sqlite-server/util"
)

func RegisterCalendarRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/calendar.ics", session.RequireAuth(db, calendarHandler(db)))
	mux.HandleFunc("/calendar/token", session.RequireAuth(db, tokenHandler(db)))
	mux.HandleFunc("/calendar/", publicCalendarHandler(db))
	mux.HandleFunc("/calendar/token/rotate", session.RequireAuth(db, rotateTokenHandler(db)))
}

func calendarHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := session.UserIDFromCtx(r.Context())
		if !ok || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Compute a simple validator: max(last_modified_epoch) and row count.
		var maxMod int64
		var n int64
		err := db.QueryRow(`
			SELECT COALESCE(MAX(last_modified_epoch),0), COUNT(*)
			FROM calendar_index WHERE user_id = ?;
		`, userID).Scan(&maxMod, &n)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		etag := `W/"` + strconv.FormatInt(maxMod, 10) + `-` + strconv.FormatInt(n, 10) + `"`
		lastMod := time.Unix(maxMod, 0).UTC().Format(http.TimeFormat)

		// Conditional GET handling
		if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if ims := r.Header.Get("If-Modified-Since"); ims != "" {
			if t, err := time.Parse(http.TimeFormat, ims); err == nil && !time.Unix(maxMod, 0).After(t) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		events, err := GetUserEvents(db, userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		ics := BuildICS(events)
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=calendar.ics")
		w.Header().Set("ETag", etag)
		w.Header().Set("Last-Modified", lastMod)
		w.Write([]byte(ics))
	}
}

func tokenHandler(db *sql.DB) http.HandlerFunc {
	type resp struct {
		Token string `json:"token"`
		// Path only; constructing absolute URL is proxy-dependent.
		UrlPath string `json:"urlPath"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := session.UserIDFromCtx(r.Context())
		if !ok || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tok, err := GetOrCreateCalendarToken(db, userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, resp{
			Token:   tok,
			UrlPath: "/api/calendar/" + tok + ".ics",
		}, http.StatusOK)
	}
}


// /api/calendar/{token}.ics  (no auth; Apple/Google won't send cookies)
func publicCalendarHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect path like: /calendar/<token>.ics
		path := r.URL.Path
		// Trim the prefix your API router applied. Here RegisterCalendarRoutes is mounted at /api,
		// so path is "/calendar/<token>.ics" already.
		const prefix = "/calendar/"
		if len(path) <= len(prefix) || path[:len(prefix)] != prefix || !strings.HasSuffix(path, ".ics") {
			http.NotFound(w, r)
			return
		}
		token := strings.TrimSuffix(path[len(prefix):], ".ics")
		if token == "" {
			http.NotFound(w, r)
			return
		}

		// Resolve token → user
		var userID string
		if err := db.QueryRow(`SELECT user_id FROM calendar_tokens WHERE token = ?`, token).Scan(&userID); err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// Update last_used_at (best effort)
		_, _ = db.Exec(`UPDATE calendar_tokens SET last_used_at = strftime('%s','now') WHERE token = ?`, token)

		// ETag/Last-Modified like the authed endpoint
		var maxMod, n int64
		if err := db.QueryRow(`
			SELECT COALESCE(MAX(last_modified_epoch),0), COUNT(*)
			FROM calendar_index WHERE user_id = ?;
		`, userID).Scan(&maxMod, &n); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		etag := `W/"` + strconv.FormatInt(maxMod, 10) + `-` + strconv.FormatInt(n, 10) + `"`
		lastMod := time.Unix(maxMod, 0).UTC().Format(http.TimeFormat)

		// Conditional GET
		if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if ims := r.Header.Get("If-Modified-Since"); ims != "" {
			if t, err := time.Parse(http.TimeFormat, ims); err == nil && !time.Unix(maxMod, 0).After(t) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		events, err := GetUserEvents(db, userID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		ics := BuildICS(events)

		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=calendar.ics")
		w.Header().Set("ETag", etag)
		w.Header().Set("Last-Modified", lastMod)
		_, _ = w.Write([]byte(ics))
	}
}

func rotateTokenHandler(db *sql.DB) http.HandlerFunc {
    type resp struct {
        Token   string `json:"token"`
        UrlPath string `json:"urlPath"`
    }
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        userID, ok := session.UserIDFromCtx(r.Context())
        if !ok || userID == "" { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

        tok, err := RotateCalendarToken(db, userID)
        if err != nil { http.Error(w, "internal error", http.StatusInternalServerError); return }

        util.WriteJSON(w, resp{
            Token:   tok,
            UrlPath: "/api/calendar/" + tok + ".ics",
        }, http.StatusOK)
    }
}
