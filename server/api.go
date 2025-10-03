package main

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strconv"
  "strings"
)

// registerRoutes wires handlers.
func registerRoutes(mux *http.ServeMux, db *sql.DB) {
  mux.HandleFunc("/", rootHandler)
  mux.HandleFunc("/books", listBooksHandler(db))
  mux.HandleFunc("/books/", getBookByIDHandler(db)) // notice the slash
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

func writeJSON(w http.ResponseWriter, v any, status int) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.WriteHeader(status)
  enc := json.NewEncoder(w)
  enc.SetIndent("", "  ")
  _ = enc.Encode(v)
}
