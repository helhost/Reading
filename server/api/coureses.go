package api

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strconv"
  "strings"

  "example.com/sqlite-server/store"
)

func RegisterCourseRoutes(mux *http.ServeMux, db *sql.DB) {
  mux.HandleFunc("/courses", requireAuth(db, coursesHandler(db)))
  mux.HandleFunc("/courses/", requireAuth(db, courseByIDHandler(db)))
}

func coursesHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
      listCoursesHandler(db)(w, r)
    case http.MethodPost:
      addCourseHandler(db)(w, r)
    default:
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
  }
}

func courseByIDHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
      getCourseByIDHandler(db)(w, r)
    case http.MethodDelete:
      deleteCourseHandler(db)(w, r)
    default:
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
  }
}

func deleteCourseHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    uid, ok := userIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/courses/"), "/")
    if len(parts[0]) == 0 {
      http.Error(w, "missing id", http.StatusBadRequest)
      return
    }
    id, err := strconv.ParseInt(parts[0], 10, 64)
    if err != nil {
      http.Error(w, "invalid id", http.StatusBadRequest)
      return
    }

    // 404 if not found / not owned
    if err := store.DeleteCourse(db, uid, id); err != nil {
      if err == sql.ErrNoRows {
        http.Error(w, "not found", http.StatusNotFound)
        return
      }
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    w.WriteHeader(http.StatusNoContent)
  }
}

// --- API DTO that ensures embedded books expose []int64 progress ---

type courseResponse struct {
  ID    int64           `json:"id"`
  Year  int64           `json:"year"`
  Term  int64           `json:"term"`
  Code  string          `json:"code"`
  Name  string          `json:"name"`
  Books []bookResponse  `json:"books,omitempty"`
}

func toCourseResponse(db *sql.DB, uid string, c store.Course) (courseResponse, error) {
  out := courseResponse{
    ID:   c.ID,
    Year: c.Year,
    Term: c.Term,
    Code: c.Code,
    Name: c.Name,
  }
  if len(c.Books) > 0 {
    out.Books = make([]bookResponse, 0, len(c.Books))
    for _, it := range c.Books {
      br, err := toBookResponse(db, uid, it)
      if err != nil {
        return courseResponse{}, err
      }
      out.Books = append(out.Books, br)
    }
  }
  return out, nil
}

func listCoursesHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    uid, ok := userIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }
    items, err := store.GetAllCourses(db, uid)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    // Map to API DTO with []int64 progress for each embedded book
    out := make([]courseResponse, 0, len(items))
    for _, c := range items {
      cr, err := toCourseResponse(db, uid, c)
      if err != nil {
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
      }
      out = append(out, cr)
    }
    writeJSON(w, out, http.StatusOK)
  }
}

type addCoursePayload struct {
  Year int64  `json:"year"`
  Term int64  `json:"term"`
  Code string `json:"code"`
  Name string `json:"name"`
}

func addCourseHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    uid, ok := userIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    var p addCoursePayload
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&p); err != nil {
      http.Error(w, "bad request", http.StatusBadRequest)
      return
    }
    if p.Year <= 0 || p.Term <= 0 || strings.TrimSpace(p.Code) == "" || strings.TrimSpace(p.Name) == "" {
      http.Error(w, "all fields are required", http.StatusBadRequest)
      return
    }

    c, err := store.AddCourse(db, uid, p.Year, p.Term, p.Code, p.Name)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }

    // Freshly created course won’t have books yet, but keep the response shape consistent
    cr, err := toCourseResponse(db, uid, c)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    writeJSON(w, cr, http.StatusCreated)
  }
}

func getCourseByIDHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    uid, ok := userIDFromCtx(r.Context())
    if !ok {
      http.Error(w, "unauthorized", http.StatusUnauthorized)
      return
    }

    parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/courses/"), "/")
    if len(parts[0]) == 0 {
      http.Error(w, "missing id", http.StatusBadRequest)
      return
    }
    id, err := strconv.ParseInt(parts[0], 10, 64)
    if err != nil {
      http.Error(w, "invalid id", http.StatusBadRequest)
      return
    }

    c, err := store.GetCourseByID(db, uid, id)
    if err != nil {
      http.Error(w, "not found", http.StatusNotFound)
      return
    }

    cr, err := toCourseResponse(db, uid, c)
    if err != nil {
      http.Error(w, "internal error", http.StatusInternalServerError)
      return
    }
    writeJSON(w, cr, http.StatusOK)
  }
}
