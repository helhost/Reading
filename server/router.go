package main

import (
	"database/sql"
	"log"
	"net/http"

	"example.com/sqlite-server/auth"
	"example.com/sqlite-server/university"
  "example.com/sqlite-server/membership"
  "example.com/sqlite-server/enrollment"
  "example.com/sqlite-server/course"
	"example.com/sqlite-server/book"
	"example.com/sqlite-server/chapter"
)

// -----------------------------------------------------------
// Router setup
// -----------------------------------------------------------

func registerRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/", rootHandler)

	auth.RegisterAuthRoutes(mux, db)
	university.RegisterUniversityRoutes(mux, db)
  membership.RegisterMembershipRoutes(mux, db)
  enrollment.RegisterEnrollmentRoutes(mux, db)
	course.RegisterCourseRoutes(mux, db)
	book.RegisterBookRoutes(mux, db)
	chapter.RegisterChapterRoutes(mux, db)

}

// -----------------------------------------------------------
// Root handler
// -----------------------------------------------------------

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("up\n"))
	log.Printf("%s %s", r.Method, r.URL.Path)
}
