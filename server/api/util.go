package api

import (
	"encoding/json"
	"net/http"
	"net/mail"
)

func writeJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func verifyEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}
