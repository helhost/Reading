package article

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

// RegisterArticleRoutes wires the /articles dispatcher behind auth.
func RegisterArticleRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/articles", session.RequireAuth(db, articlesHandler(db)))
}

// Dispatcher for /articles (auth already applied by middleware).
func articlesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postArticleHandler(db)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// POST /articles
// Body: { "courseId": number, "title": string, "author": string, "location": string? }
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
