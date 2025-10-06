package util

import (
	"encoding/json"
  "net/http"
  "net/mail"
  "os"
  "strconv"
  "strings"
)

func WriteJSON(w http.ResponseWriter, v any, status int) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.WriteHeader(status)
  enc := json.NewEncoder(w)
  enc.SetIndent("", "  ")
  _ = enc.Encode(v)
}

func VerifyEmail(s string) bool {
  _, err := mail.ParseAddress(s)
  return err == nil
}

func IsProd() bool {
  v := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
  return v == "prod"
}

// ParseInt64Query extracts and parses a positive int64 query parameter.
func ParseInt64Query(r *http.Request, key string) (int64, error) {
  v := strings.TrimSpace(r.URL.Query().Get(key))
  if v == "" {
    return 0, strconv.ErrSyntax
  }
  n, err := strconv.ParseInt(v, 10, 64)
  if err != nil {
    return 0, err
  }
  return n, nil
}
