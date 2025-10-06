
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/sqlite-server/auth"
	"example.com/sqlite-server/university"
)

// -----------------------------------------------------------
// CORS configuration and middleware
// -----------------------------------------------------------

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

		w.Header().Set("Vary", "Origin")

		if allowed {
			if wildcard {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// -----------------------------------------------------------
// Router setup
// -----------------------------------------------------------

func registerRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/", rootHandler)

	auth.RegisterAuthRoutes(mux, db)
  university.RegisterUniversityRoutes(mux, db)
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
