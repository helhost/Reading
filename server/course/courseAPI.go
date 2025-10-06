package course

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"example.com/sqlite-server/membership"
	"example.com/sqlite-server/session"
	"example.com/sqlite-server/util"
)

// RegisterCourseRoutes wires the course endpoints.
func RegisterCourseRoutes(mux *http.ServeMux, db *sql.DB) {
	// Auth required for both GET (my courses) and POST (create)
	mux.HandleFunc("/courses", session.RequireAuth(db, coursesHandler(db)))
}

// Dispatcher for /courses (auth already applied by middleware).
func coursesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getMyCoursesForUniversityHandler(db)(w, r)
		case http.MethodPost:
			postCourseHandler(db)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// GET /courses?universityId=UUID
// Returns ONLY the caller's enrolled courses for the given university.
// Authorization: user must be a member of the university (defense-in-depth).
func getMyCoursesForUniversityHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := session.UserIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		uniID := strings.TrimSpace(r.URL.Query().Get("universityId"))
		if uniID == "" {
			http.Error(w, "universityId is required", http.StatusBadRequest)
			return
		}

		// Membership gate (enrollment already implies membership, but keep this for clarity)
		isMember, err := membership.IsMember(db, uid, uniID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !isMember {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		list, err := ListMyCoursesByUniversity(db, uid, uniID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, list, http.StatusOK)
	}
}

// POST /courses
// Body: { "universityId": "uuid", "year": 2025, "term": 1, "code": "CS101", "name": "Intro to CS" }
// Requires: caller is a member of universityId (create shared course for the university).
func postCourseHandler(db *sql.DB) http.HandlerFunc {
	type payload struct {
		UniversityID string `json:"universityId"`
		Year         int64  `json:"year"`
		Term         int64  `json:"term"`
		Code         string `json:"code"`
		Name         string `json:"name"`
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

		p.UniversityID = strings.TrimSpace(p.UniversityID)
		p.Code = strings.TrimSpace(p.Code)
		p.Name = strings.TrimSpace(p.Name)
		if p.UniversityID == "" || p.Year <= 0 || p.Term < 1 || p.Term > 4 || p.Code == "" || p.Name == "" {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		isMember, err := membership.IsMember(db, uid, p.UniversityID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !isMember {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		c, err := AddCourse(db, p.UniversityID, p.Year, p.Term, p.Code, p.Name)
		if err != nil {
			lc := strings.ToLower(err.Error())
			switch {
			case strings.Contains(lc, "university not found"):
				http.Error(w, "university not found", http.StatusBadRequest)
				return
			case strings.Contains(lc, "course already exists"):
				http.Error(w, "course already exists", http.StatusConflict)
				return
			case strings.Contains(lc, "invalid input"):
				http.Error(w, "invalid input", http.StatusBadRequest)
				return
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		util.WriteJSON(w, c, http.StatusCreated)
	}
}
