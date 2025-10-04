package main

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "os"
  "strconv"
  "strings"
)

// parse ALLOW_ORIGIN once at startup
var allowedOrigins = func() map[string]struct{} {
    raw := os.Getenv("ALLOW_ORIGIN")
    m := make(map[string]struct{})
    for _, s := range strings.Split(raw, ",") {
        s = strings.TrimSpace(s)
        if s != "" {
            m[s] = struct{}{}
        }
    }
    // default for local dev if unset
    if len(m) == 0 {
        m["http://localhost:5173"] = struct{}{}
    }
    return m
}()

func originAllowed(o string) (bool, bool) {
    if o == "" {
        return false, false
    }
    if _, ok := allowedOrigins["*"]; ok {
        return true, true // (allowed, wildcard)
    }
    _, ok := allowedOrigins[o]
    return ok, false
}

func withCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqOrigin := r.Header.Get("Origin")
        allowed, wildcard := originAllowed(reqOrigin)

        // Always signal that response may vary by Origin
        w.Header().Set("Vary", "Origin")

        if allowed {
            if wildcard {
                w.Header().Set("Access-Control-Allow-Origin", "*")
            } else {
                w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
            }
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            // If you later use cookies/session: also set
            // w.Header().Set("Access-Control-Allow-Credentials", "true")
        }

        if r.Method == http.MethodOptions {
            // Preflight: no body needed
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func registerRoutes(mux *http.ServeMux, db *sql.DB) {
    mux.HandleFunc("/", rootHandler)

    // books routes
    mux.HandleFunc("/books", booksHandler(db))
    mux.HandleFunc("/books/", bookByIDHandler(db))

    // courses routes
    mux.HandleFunc("/courses", coursesHandler(db))
    mux.HandleFunc("/courses/", courseByIDHandler(db))
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
  CourseID    *int64  `json:"courseId,omitempty"`
}

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

    item, err := AddBook(db, p.Title, p.Author, p.NumChapters, p.Link, p.CourseID)
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


// deleteBookHandler handles DELETE /books/{id}.
func deleteBookHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
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

		// Try delete
		err = DeleteBook(db, id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Just return 204 No Content
		w.WriteHeader(http.StatusNoContent)
	}
}



// courses

// /courses: GET list, POST create
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

// /courses/{id}: GET single (you can add PATCH/DELETE later)
func courseByIDHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            getCourseByIDHandler(db)(w, r)
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}

func listCoursesHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        items, err := GetAllCourses(db)
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

        c, err := AddCourse(db, p.Year, p.Term, p.Code, p.Name)
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

        c, err := GetCourseByID(db, id)
        if err != nil {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        writeJSON(w, c, http.StatusOK)
    }
}


func writeJSON(w http.ResponseWriter, v any, status int) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.WriteHeader(status)
  enc := json.NewEncoder(w)
  enc.SetIndent("", "  ")
  _ = enc.Encode(v)
}
