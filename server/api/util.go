package api

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"os"
	"strings"
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


func isProd() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	return v == "prod"
}
