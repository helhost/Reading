
package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"example.com/sqlite-server/store"
)

// handleBookProgressPATCH handles PATCH /books/{id}/progress
// It assumes auth + bookID parsing happened upstream.
func handleBookProgressPATCH(db *sql.DB, uid string, bookID int64, w http.ResponseWriter, r *http.Request) {
	var p struct {
		Chapter int64  `json:"chapter"`
		Action  string `json:"action"` // "add" | "remove"
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil || p.Chapter <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var err error
	switch strings.ToLower(p.Action) {
	case "add":
		err = store.AddProgress(db, uid, bookID, p.Chapter)
	case "remove":
		err = store.RemoveProgress(db, uid, bookID, p.Chapter)
	default:
		http.Error(w, "invalid action (use 'add' or 'remove')", http.StatusBadRequest)
		return
	}
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	list, err := store.ListProgress(db, uid, bookID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, list, http.StatusOK)
}
