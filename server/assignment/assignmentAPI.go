package assignment

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/sqlite-server/enrollment"
	"example.com/sqlite-server/membership"
	"example.com/sqlite-server/session"
	"example.com/sqlite-server/util"
)

// RegisterAssignmentRoutes wires the /assignments endpoints behind auth.
func RegisterAssignmentRoutes(mux *http.ServeMux, db *sql.DB) {
	// Collection endpoints
	mux.HandleFunc("/assignments", session.RequireAuth(db, assignmentsHandler(db)))
	// Item subroutes (e.g., /assignments/{id}/deadline)
	mux.HandleFunc("/assignments/", session.RequireAuth(db, assignmentsDispatcher(db)))
}

// Dispatcher for /assignments (collection)
func assignmentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postAssignmentHandler(db)(w, r)
		case http.MethodGet:
			getAssignmentsForCourseHandler(db)(w, r)
		case http.MethodDelete:
			deleteAssignmentHandler(db)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// Dispatcher for /assignments/{id}/...
func assignmentsDispatcher(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect: /assignments/{id}/(deadline|progress)
		path := strings.TrimPrefix(r.URL.Path, "/assignments/")
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
				patchAssignmentDeadlineHandler(db, id)(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		case "progress":
			switch r.Method {
			case http.MethodPatch:
				patchAssignmentProgressHandler(db, id)(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}

// POST /assignments
// Body: { "courseId": number, "title": string, "description"?: string }
// Auth: caller must be a member of the university that owns the course.
// Behavior: creates assignment with deadline = NULL.
func postAssignmentHandler(db *sql.DB) http.HandlerFunc {
	type payload struct {
		CourseID    int64   `json:"courseId"`
		Title       string  `json:"title"`
		Description *string `json:"description,omitempty"`
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
		if p.Description != nil {
			s := strings.TrimSpace(*p.Description)
			if s == "" {
				p.Description = nil
			} else {
				p.Description = &s
			}
		}
		if p.CourseID <= 0 || p.Title == "" {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		// Membership gate: user must belong to the university that owns the course.
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

		a, err := AddAssignment(db, p.CourseID, p.Title, p.Description)
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

		util.WriteJSON(w, a, http.StatusCreated)
	}
}


// GET /assignments?courseId=123
// Auth: caller must be ENROLLED in the course.
// Returns: []AssignmentWithStatus including per-user "completed" flag.
func getAssignmentsForCourseHandler(db *sql.DB) http.HandlerFunc {
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

		enrolled, err := enrollment.UserEnrolledInCourse(db, uid, courseID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		list, err := ListAssignmentsByCourseWithProgress(db, courseID, uid)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, list, http.StatusOK)
	}
}

// PATCH /assignments/{id}/deadline
// Body: { "deadline": number|null }  (unix seconds; null clears)
// Auth: caller must be enrolled in the assignment's course.
// Returns: 200 OK with updated Assignment
func patchAssignmentDeadlineHandler(db *sql.DB, assignmentID int64) http.HandlerFunc {
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

		// Ensure it exists (404 semantics).
		if _, err := AssignmentUniversityID(db, assignmentID); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Must be enrolled in owning course.
		canEdit, err := UserEnrolledInAssignmentCourse(db, uid, assignmentID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !canEdit {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Update
		if err := SetAssignmentDeadline(db, assignmentID, p.Deadline); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Return updated resource
		a, err := GetAssignment(db, assignmentID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, a, http.StatusOK)
	}
}


// PATCH /assignments/{id}/progress
// Body: { "completed": boolean }
// Auth: caller must be enrolled in the assignment's course.
// Returns: 200 OK with { "completed": true|false }
func patchAssignmentProgressHandler(db *sql.DB, assignmentID int64) http.HandlerFunc {
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

		// Ensure it exists (404 semantics).
		if _, err := AssignmentUniversityID(db, assignmentID); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Must be enrolled in owning course.
		canEdit, err := UserEnrolledInAssignmentCourse(db, uid, assignmentID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !canEdit {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Apply change (service layer).
		if err := SetAssignmentProgress(db, uid, assignmentID, *p.Completed); err != nil {
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


// DELETE /assignments
// Body: { "assignmentId": number }
// Auth: must be enrolled in the assignment's course.
// 204 if deleted; 404 if not found; 409 if any user has completed it.
func deleteAssignmentHandler(db *sql.DB) http.HandlerFunc {
	type payload struct {
		AssignmentID int64 `json:"assignmentId"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var p payload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil || p.AssignmentID <= 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		uid, ok := session.UserIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Must be enrolled in owning course
		enrolled, err := UserEnrolledInAssignmentCourse(db, uid, p.AssignmentID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		deleted, derr := DeleteAssignmentIfNoProgress(db, p.AssignmentID)
		if derr != nil {
			switch {
			case derr == sql.ErrNoRows:
				http.Error(w, "not found", http.StatusNotFound)
				return
			case strings.Contains(strings.ToLower(derr.Error()), "invalid input"):
				http.Error(w, "invalid input", http.StatusBadRequest)
				return
			case strings.Contains(strings.ToLower(derr.Error()), "has progress"):
				http.Error(w, "conflict: assignment has progress", http.StatusConflict)
				return
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}
		if !deleted {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
