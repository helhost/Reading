package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"example.com/sqlite-server/store"
	"example.com/sqlite-server/middleware"
)

func main() {
	// 1. Database setup
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db" // default for local dev
	}

	dsn := "file:" + dbPath + "?_pragma=busy_timeout(5000)"
	db, err := store.OpenDB(dsn)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer db.Close()

	if err := store.EnsureSchema(db); err != nil { log.Fatal("Failed to load schema", err) }
	if err := store.EnsureCalendar(db); err != nil { log.Fatal("Failed to load calendar schema", err) }

	// 2. API routes
	apiMux := http.NewServeMux()
	registerRoutes(apiMux, db)

	// 3. Top-level mux
	mux := http.NewServeMux()

	// Mount API under /api with CORS middleware
	mux.Handle("/api/", http.StripPrefix("/api", middleware.WithCORS(apiMux)))

	// Serve static client (if present)
	fs := http.FileServer(http.Dir("./client"))
	mux.Handle("/", fs)

	// 4. HTTP server configuration
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
