package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	// open DB + ensure schema
	db, err := openDB("file:data.db?_pragma=busy_timeout(5000)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		log.Fatal(err)
	}

	// wire routes
	mux := http.NewServeMux()
	registerRoutes(mux, db)

	// start server
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("listening on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
