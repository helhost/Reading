package middleware

import (
	"net/http"
	"os"
	"strings"
)

var AllowedOrigins = func() map[string]struct{} {
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

func OriginAllowed(o string) (bool, bool) {
	if o == "" {
		return false, false
	}
	if _, ok := AllowedOrigins["*"]; ok {
		return true, true // (allowed, wildcard)
	}
	_, ok := AllowedOrigins[o]
	return ok, false
}

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")
		allowed, wildcard := OriginAllowed(reqOrigin)

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
