package chapter

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/sqlite-server/session"
)

// RegisterChapterRoutes wires chapter endpoints.
func RegisterChapterRoutes(mux *http.ServeMux, db *sql.DB) {
	// Only one endpoint for set/clear deadline. Auth required.
	mux.HandleFunc("/chapters/", session.RequireAuth(db, chaptersDispatcher(db)))
}

func chaptersDispatcher(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect: /chapters/{id}/deadline  (PATCH only)
		path := strings.TrimPrefix(r.URL.Path, "/chapters/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] != "deadline" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPatch:
			patchChapterDeadlineHandler(db, id)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func patchChapterDeadlineHandler(db *sql.DB, chapterID int64) http.HandlerFunc {
	type payload struct {
		Deadline *int64 `json:"deadline"` // unix seconds; nil clears deadline
	}
	const maxDeadline = int64(4102444800) // 2100-01-01T00:00:00Z

	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := session.UserIDFromCtx(r.Context())
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
		if p.Deadline != nil {
			if *p.Deadline < 0 || *p.Deadline > maxDeadline {
				http.Error(w, "invalid deadline", http.StatusBadRequest)
				return
			}
		}

		// Ensure the chapter exists (optional but yields 404 when missing).
		if _, err := ChapterUniversityID(db, chapterID); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Stricter access control: must be enrolled in the course owning this chapter.
		canEdit, err := UserEnrolledInChapterCourse(db, uid, chapterID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !canEdit {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Apply the deadline change.
		if err := SetChapterDeadline(db, chapterID, p.Deadline); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
