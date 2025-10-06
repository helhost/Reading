package article

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

// RegisterArticleRoutes wires the /articles endpoints behind auth.
func RegisterArticleRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/articles", session.RequireAuth(db, articlesHandler(db)))
	mux.HandleFunc("/articles/", session.RequireAuth(db, articlesDispatcher(db)))
}

// Dispatcher for /articles (collection)
func articlesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postArticleHandler(db)(w, r)
		case http.MethodGet:
			getArticlesForCourseHandler(db)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// Dispatcher for /articles/{id}/...
func articlesDispatcher(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect: /articles/{id}/deadline  (PATCH only)
		path := strings.TrimPrefix(r.URL.Path, "/articles/")
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
			patchArticleDeadlineHandler(db, id)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// POST /articles
// Body: { "courseId": number, "title": string, "author": string, "location"?: string }
// Auth: caller must be a member of the university that owns the course.
// Behavior: creates article with deadline = NULL.
func postArticleHandler(db *sql.DB) http.HandlerFunc {
	type payload struct {
		CourseID int64   `json:"courseId"`
		Title    string  `json:"title"`
		Author   string  `json:"author"`
		Location *string `json:"location,omitempty"`
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
		if p.Location != nil {
			s := strings.TrimSpace(*p.Location)
			if s == "" {
				p.Location = nil
			} else {
				p.Location = &s
			}
		}
		if p.CourseID <= 0 || p.Title == "" || p.Author == "" {
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

		a, err := AddArticle(db, p.CourseID, p.Title, p.Author, p.Location)
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

// PATCH /articles/{id}/deadline
// Body: { "deadline": number|null }  (unix seconds; null clears)
func patchArticleDeadlineHandler(db *sql.DB, articleID int64) http.HandlerFunc {
	type payload struct {
		Deadline *int64 `json:"deadline"`
	}
	const maxDeadline = int64(4102444800)

	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := session.UserIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var p payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if p.Deadline != nil && (*p.Deadline < 0 || *p.Deadline > maxDeadline) {
			http.Error(w, "invalid deadline", http.StatusBadRequest)
			return
		}

		// Access checks
		if _, err := ArticleUniversityID(db, articleID); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		canEdit, err := UserEnrolledInArticleCourse(db, uid, articleID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !canEdit {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Update deadline
		if err := SetArticleDeadline(db, articleID, p.Deadline); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Fetch and return updated article
		a, err := GetArticle(db, articleID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, a, http.StatusOK)
	}
}

func getArticlesForCourseHandler(db *sql.DB) http.HandlerFunc {
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

		uniID, err := enrollment.CourseUniversity(db, courseID)
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

		list, err := ListArticlesByCourse(db, courseID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, list, http.StatusOK)
	}
}
