package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db" // fallback for local
	}

	dsn := "file:" + dbPath + "?_pragma=busy_timeout(5000)"
	db, err := openDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		log.Fatal(err)
	}

	// API routes
	apiMux := http.NewServeMux()
	registerRoutes(apiMux, db)

	// top-level mux
	mux := http.NewServeMux()

	// mount API
	mux.Handle("/api/", http.StripPrefix("/api", withCORS(apiMux)))

	// serve static files (client)
	fs := http.FileServer(http.Dir("./client"))
	mux.Handle("/", fs)


	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
