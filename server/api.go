package main

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strconv"
  "strings"
)

func registerRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/books", booksHandler(db))
	mux.HandleFunc("/books/", bookByIDHandler(db))
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
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodGet {
    http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    return
  }
  w.Header().Set("Content-Type", "text/plain; charset=utf-8")
  w.Write([]byte("up\n"))
}

func listBooksHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    items, err := GetAllItems(db)
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

    // path looks like: /books/{id}
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

    book, err := GetBookByID(db, id)
    if err != nil {
      http.Error(w, "not found", http.StatusNotFound)
      return
    }
    writeJSON(w, book, http.StatusOK)
  }
}


// addBookPayload is the expected JSON body for creating a book.
type addBookPayload struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	NumChapters int64   `json:"numChapters"`
	Link        *string `json:"link,omitempty"`
}

// addBookHandler handles POST /books with a JSON body.
func addBookHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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

		item, err := AddBook(db, p.Title, p.Author, p.NumChapters, p.Link)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, item, http.StatusCreated)
	}
}

type patchBookPayload struct {
	CompletedChapters int64 `json:"completedChapters"`
}


// patchBookHandler handles PATCH /books/{id}.
func patchBookHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract ID
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

		// Decode body
		var p patchBookPayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Validate against numChapters
		book, err := GetBookByID(db, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if p.CompletedChapters < 0 || p.CompletedChapters > book.NumChapters {
			http.Error(w, "completedChapters must be between 0 and numChapters", http.StatusBadRequest)
			return
		}

		// Update
		updated, err := UpdateCompletedChapters(db, id, p.CompletedChapters)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, updated, http.StatusOK)
	}
}


func writeJSON(w http.ResponseWriter, v any, status int) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.WriteHeader(status)
  enc := json.NewEncoder(w)
  enc.SetIndent("", "  ")
  _ = enc.Encode(v)
}
