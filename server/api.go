package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// registerRoutes wires: GET / -> return all items as JSON.
func registerRoutes(mux *http.ServeMux, db *sql.DB) {

	// get all
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		items, err := GetAllItems(db)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(items)
	})

}
