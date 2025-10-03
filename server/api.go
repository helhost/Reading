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
  mux.HandleFunc("/books", booksHandler(db))   // <-- was listBooksHandler
  mux.HandleFunc("/books/", getBookByIDHandler(db))
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

func writeJSON(w http.ResponseWriter, v any, status int) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.WriteHeader(status)
  enc := json.NewEncoder(w)
  enc.SetIndent("", "  ")
  _ = enc.Encode(v)
}
