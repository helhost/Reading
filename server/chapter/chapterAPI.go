package chapter

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/sqlite-server/session"
	"example.com/sqlite-server/util"
)

// RegisterChapterRoutes wires chapter endpoints.
func RegisterChapterRoutes(mux *http.ServeMux, db *sql.DB) {
	// Only one endpoint for set/clear deadline. Auth required.
	mux.HandleFunc("/chapters/", session.RequireAuth(db, chaptersDispatcher(db)))
}

func chaptersDispatcher(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect: /chapters/{id}/(deadline|progress)
		path := strings.TrimPrefix(r.URL.Path, "/chapters/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[0] == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		switch parts[1] {
		case "deadline":
			switch r.Method {
			case http.MethodPatch:
				patchChapterDeadlineHandler(db, id)(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		case "progress":
			switch r.Method {
			case http.MethodPatch:
				patchChapterProgressHandler(db, id)(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}


// PATCH /chapters/{id}/deadline
// Body: { "deadline": number|null }  (unix seconds; null clears)
// Auth: caller must be enrolled in the chapter's course.
func patchChapterDeadlineHandler(db *sql.DB, chapterID int64) http.HandlerFunc {
	type payload struct {
		Deadline *int64 `json:"deadline"`
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

		// Ensure the chapter exists (for 404 semantics).
		if _, err := ChapterUniversityID(db, chapterID); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Must be enrolled in the owning course.
		canEdit, err := UserEnrolledInChapterCourse(db, uid, chapterID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !canEdit {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Apply update.
		if err := SetChapterDeadline(db, chapterID, p.Deadline); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Reload and return updated resource from the service layer.
		c, err := GetChapter(db, chapterID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, c, http.StatusOK)
	}
}


func patchChapterProgressHandler(db *sql.DB, chapterID int64) http.HandlerFunc {
	type payload struct {
		Completed *bool `json:"completed"`
	}
	type resp struct {
		Completed bool `json:"completed"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := session.UserIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var p payload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil || p.Completed == nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		canEdit, err := UserEnrolledInChapterCourse(db, uid, chapterID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !canEdit {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := SetChapterProgress(db, uid, chapterID, *p.Completed); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, resp{Completed: *p.Completed}, http.StatusOK)
	}
}
