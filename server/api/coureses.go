package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/sqlite-server/store"
)

func coursesHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            listCoursesHandler(db)(w, r)
        case http.MethodPost:
            addCourseHandler(db)(w, r)
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}

func courseByIDHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            getCourseByIDHandler(db)(w, r)
        case http.MethodDelete:
            deleteCourseHandler(db)(w, r)
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}


func deleteCourseHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodDelete {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Extract ID from /courses/{id}
        parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/courses/"), "/")
        if len(parts[0]) == 0 {
            http.Error(w, "missing id", http.StatusBadRequest)
            return
        }
        id, err := strconv.ParseInt(parts[0], 10, 64)
        if err != nil {
            http.Error(w, "invalid id", http.StatusBadRequest)
            return
        }

        // Optional existence check -> 404 if not found
        if _, err := store.GetCourseByID(db, id); err != nil {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }

        // Delete (books will have course_id set to NULL via FK ON DELETE SET NULL)
        if err := store.DeleteCourse(db, id); err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}


func listCoursesHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        items, err := store.GetAllCourses(db)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        writeJSON(w, items, http.StatusOK)
    }
}


type addCoursePayload struct {
    Year int64  `json:"year"`
    Term int64  `json:"term"`
    Code string `json:"code"`
    Name string `json:"name"`
}

func addCourseHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var p addCoursePayload
        dec := json.NewDecoder(r.Body)
        dec.DisallowUnknownFields()
        if err := dec.Decode(&p); err != nil {
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }
        if p.Year <= 0 || p.Term <= 0 || strings.TrimSpace(p.Code) == "" || strings.TrimSpace(p.Name) == "" {
            http.Error(w, "all fields are required", http.StatusBadRequest)
            return
        }

        c, err := store.AddCourse(db, p.Year, p.Term, p.Code, p.Name)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        writeJSON(w, c, http.StatusCreated)
    }
}


func getCourseByIDHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // path: /courses/{id}
        parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/courses/"), "/")
        if len(parts[0]) == 0 {
            http.Error(w, "missing id", http.StatusBadRequest)
            return
        }
        id, err := strconv.ParseInt(parts[0], 10, 64)
        if err != nil {
            http.Error(w, "invalid id", http.StatusBadRequest)
            return
        }

        c, err := store.GetCourseByID(db, id)
        if err != nil {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        writeJSON(w, c, http.StatusOK)
    }
}



func RegisterCourseRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/courses", coursesHandler(db))
	mux.HandleFunc("/courses/", courseByIDHandler(db))
}
