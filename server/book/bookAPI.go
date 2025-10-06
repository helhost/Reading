package book

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"example.com/sqlite-server/enrollment"
	"example.com/sqlite-server/membership"
	"example.com/sqlite-server/session"
	"example.com/sqlite-server/util"
)

func RegisterBookRoutes(mux *http.ServeMux, db *sql.DB) {
	// Both endpoints require auth and membership to the course's university.
	mux.HandleFunc("/books", session.RequireAuth(db, booksHandler(db)))
}

func booksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getBooksForCourseHandler(db)(w, r)
		case http.MethodPost:
			postBookHandler(db)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}


// GET /books?courseId=123
// Returns books for the course, each with embedded chapters and per-user "completed".
// Access: caller must be ENROLLED in the course (not just university member).
func getBooksForCourseHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := session.UserIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		courseID, err := util.ParseInt64Query(r, "courseId")
		if err != nil || courseID <= 0 {
			http.Error(w, "courseId is required", http.StatusBadRequest)
			return
		}

		// Tighten access: must be enrolled
		enrolled, err := enrollment.UserEnrolledInCourse(db, uid, courseID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		list, err := ListBooksByCourseWithProgress(db, courseID, uid)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, list, http.StatusOK)
	}
}

// POST /books
// Body:
// { "courseId": 1, "title": "...", "author": "...", "numChapters": 10, "location": "Library shelf 3A" }
// Returns created book with embedded chapters.
func postBookHandler(db *sql.DB) http.HandlerFunc {
	type payload struct {
		CourseID    int64   `json:"courseId"`
		Title       string  `json:"title"`
		Author      string  `json:"author"`
		NumChapters *int64  `json:"numChapters,omitempty"`
		Location    *string `json:"location,omitempty"`
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
		if err := dec.Decode(&p); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		p.Title = strings.TrimSpace(p.Title)
		p.Author = strings.TrimSpace(p.Author)
		if p.CourseID <= 0 || p.Title == "" || p.Author == "" {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		uniID, err := enrollment.CourseUniversity(db, p.CourseID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "course not found", http.StatusBadRequest)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		isMember, err := membership.IsMember(db, uid, uniID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !isMember {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		b, err := AddBook(db, p.CourseID, p.Title, p.Author, p.NumChapters, p.Location)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "invalid input":
				http.Error(w, "invalid input", http.StatusBadRequest)
				return
			case "course not found":
				http.Error(w, "course not found", http.StatusBadRequest)
				return
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}
		util.WriteJSON(w, b, http.StatusCreated)
	}
}
