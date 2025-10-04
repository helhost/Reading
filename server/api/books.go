package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/sqlite-server/store"
)

func RegisterBookRoutes(mux *http.ServeMux, db *sql.DB) {
	// All book routes are user-scoped, so require auth.
	mux.HandleFunc("/books", requireAuth(db, booksHandler(db)))
	mux.HandleFunc("/books/", requireAuth(db, bookByIDHandler(db)))
}

func booksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listBooksHandler(db)(w, r)
		case http.MethodPost:
			addBookHandler(db)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func bookByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getBookByIDHandler(db)(w, r)
		case http.MethodPatch:
			patchBookHandler(db)(w, r)
		case http.MethodDelete:
			deleteBookHandler(db)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listBooksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uid, ok := userIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		items, err := store.GetAllItems(db, uid)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, items, http.StatusOK)
	}
}

func getBookByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uid, ok := userIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/books/"), "/")
		if len(parts[0]) == 0 {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		book, err := store.GetBookByID(db, uid, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, book, http.StatusOK)
	}
}

type addBookPayload struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	NumChapters int64   `json:"numChapters"`
	Link        *string `json:"link,omitempty"`
	CourseID    *int64  `json:"courseId,omitempty"`
}

func addBookHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uid, ok := userIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var p addBookPayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(p.Title) == "" || strings.TrimSpace(p.Author) == "" {
			http.Error(w, "title and author are required", http.StatusBadRequest)
			return
		}
		if p.NumChapters < 0 {
			http.Error(w, "numChapters must be >= 0", http.StatusBadRequest)
			return
		}

		item, err := store.AddBook(db, uid, p.Title, p.Author, p.NumChapters, p.Link, p.CourseID)
		if err != nil {
			if err == sql.ErrNoRows && p.CourseID != nil {
				http.Error(w, "course not found", http.StatusBadRequest)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, item, http.StatusCreated)
	}
}

type patchBookPayload struct {
	CompletedChapters int64 `json:"completedChapters"`
}

func patchBookHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uid, ok := userIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/books/"), "/")
		if len(parts[0]) == 0 {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var p patchBookPayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		book, err := store.GetBookByID(db, uid, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if p.CompletedChapters < 0 || p.CompletedChapters > book.NumChapters {
			http.Error(w, "completedChapters must be between 0 and numChapters", http.StatusBadRequest)
			return
		}

		updated, err := store.UpdateCompletedChapters(db, uid, id, p.CompletedChapters)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, updated, http.StatusOK)
	}
}

func deleteBookHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uid, ok := userIDFromCtx(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/books/"), "/")
		if len(parts[0]) == 0 {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := store.DeleteBook(db, uid, id); err != nil {
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
