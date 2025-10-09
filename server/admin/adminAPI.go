package admin

import (
	"database/sql"
	"net/http"

	"example.com/sqlite-server/session"
	"example.com/sqlite-server/util"
)

type countResp struct {
	Count int64 `json:"count"`
}

func RegisterAdminRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/admin/users/count",
		session.RequireAuth(db, adminOnly(db, usersCountHandler(db))),
	)
}

func adminOnly(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := session.UserIDFromCtx(r.Context())
		if !ok || uid == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		okAdmin, err := IsAdmin(db, uid)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !okAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func usersCountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		n, err := CountUsers(db)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, countResp{Count: n}, http.StatusOK)
	}
}
