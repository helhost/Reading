# API Documentation

All requests and responses use JSON with `Content-Type: application/json`.  
Authentication is via an HTTP-only session cookie set by login/register.

Common error codes:
- 400 — Bad request
- 401 — Unauthorized (missing/expired session)
- 403 — Forbidden (membership/enrollment required)
- 404 — Not found
- 409 — Conflict (progress or dependency prevents deletion)
- 500 — Internal error

---

## AUTH

### POST /api/register
Create a new user and log in.

Request:
```json
{ "email": "user@example.com", "password": "atLeast8Chars" }
```

Response (201 Created):
```json
{ "userId": "uuid-string" }
```

---

### POST /api/login
Log in existing user.

Request:
```json
{ "email": "user@example.com", "password": "plaintext" }
```

Response (200 OK):
```json
{ "userId": "uuid-string" }
```

---

### POST /api/logout
Invalidate current session.

Response: 204 No Content

---

### GET /api/me
Get current user info.

Response (200 OK):
```json
{ "userId": "uuid-string", "email": "user@example.com" }
```

---

## UNIVERSITIES

### GET /api/universities
List all universities (public).

Response (200 OK):
```json
[
  { "id": "uuid-string", "name": "University Name" }
]
```

---

### POST /api/universities
Create a university (auth required).

Request:
```json
{ "name": "University Name" }
```

Response (201 Created):
```json
{ "id": "uuid-string", "name": "University Name" }
```

---

### DELETE /api/universities
Delete a university if empty (auth required).

Request:
```json
{ "universityId": "uuid-string" }
```

Response: 204 No Content

---

## MEMBERSHIPS (User ↔ University) — Auth Required

### GET /api/user-universities
List universities the user is a member of.

Response (200 OK):
```json
[
  { "userId": "uuid-string", "universityId": "uuid-string" }
]
```

---

### POST /api/user-universities
Join a university (idempotent).

Request:
```json
{ "universityId": "uuid-string" }
```

Response (201 Created or 200 OK):
```json
{ "userId": "uuid-string", "universityId": "uuid-string" }
```

---

### DELETE /api/user-universities
Leave a university (idempotent).

Request:
```json
{ "universityId": "uuid-string" }
```

Response: 204 No Content

---

## COURSES

### GET /api/courses
List my enrolled courses for a university (auth + membership required).

Query: `?universityId=uuid`

Response (200 OK):
```json
[
  { "id": 1, "universityId": "uuid", "year": 2025, "term": 1, "code": "CS101", "name": "Intro to CS" }
]
```

---

### GET /api/course-catalog
List all courses in a university (auth + membership required).

Query: `?universityId=uuid`

Response (200 OK):
```json
[
  { "id": 1, "universityId": "uuid", "year": 2025, "term": 1, "code": "CS101", "name": "Intro to CS" }
]
```

---

### POST /api/courses
Create a course (auth + membership required).

Request:
```json
{
  "universityId": "uuid",
  "year": 2025,
  "term": 1,
  "code": "CS101",
  "name": "Intro to CS"
}
```

Response (201 Created):
```json
{ "id": 1, "universityId": "uuid", "year": 2025, "term": 1, "code": "CS101", "name": "Intro to CS" }
```

---

### DELETE /api/courses
Delete a course (auth + membership required). Only if it has no books, articles, or assignments.

Request:
```json
{ "courseId": 1 }
```

Response: 204 No Content

---

## ENROLLMENTS (User ↔ Course) — Auth Required

### POST /api/user-courses
Enroll in a course (idempotent).

Request:
```json
{ "courseId": 123 }
```

Response (201 Created or 200 OK):
```json
{ "userId": "uuid-string", "courseId": 123 }
```

---

### DELETE /api/user-courses
Unenroll from a course (idempotent).

Request:
```json
{ "courseId": 123 }
```

Response: 204 No Content

---

## BOOKS — Auth Required

### GET /api/books
List books for a course (enrolled users only).

Query: `?courseId=123`

Response (200 OK):
```json
[
  {
    "id": 1,
    "title": "Book Title",
    "author": "Author",
    "numChapters": 10,
    "location": "Shelf 3A",
    "completed": false
  }
]
```

---

### POST /api/books
Create a book (membership in owning university required).

Request:
```json
{
  "courseId": 123,
  "title": "Book Title",
  "author": "Author Name",
  "numChapters": 10,
  "location": "Library shelf 3A"
}
```

Response (201 Created):
```json
{ "id": 1, "courseId": 123, "title": "Book Title", "author": "Author Name" }
```

---

### DELETE /api/books
Delete a book (enrolled users only). Fails if any chapter has progress.

Request:
```json
{ "bookId": 1 }
```

Response: 204 No Content

---

## CHAPTERS — Auth Required

### PATCH /api/chapters/{id}/deadline
Set or clear chapter deadline.

Request:
```json
{ "deadline": 1735689600 }  // or null
```

Response (200 OK):
```json
{ "id": 5, "deadline": 1735689600 }
```

---

### PATCH /api/chapters/{id}/progress
Mark chapter as completed/incomplete.

Request:
```json
{ "completed": true }
```

Response (200 OK):
```json
{ "completed": true }
```

---

## ARTICLES — Auth Required

### POST /api/articles
Create article (membership in owning university required).

Request:
```json
{ "courseId": 123, "title": "Article Title", "author": "Author Name", "location": "Optional" }
```

Response (201 Created):
```json
{ "id": 1, "courseId": 123, "title": "Article Title", "author": "Author Name" }
```

---

### GET /api/articles
List articles for a course (enrolled users only).

Query: `?courseId=123`

Response (200 OK):
```json
[
  { "id": 1, "title": "Article Title", "completed": false }
]
```

---

### PATCH /api/articles/{id}/deadline
Set or clear article deadline.

Request:
```json
{ "deadline": 1735689600 }
```

Response (200 OK):
```json
{ "id": 1, "deadline": 1735689600 }
```

---

### PATCH /api/articles/{id}/progress
Mark article progress.

Request:
```json
{ "completed": true }
```

Response (200 OK):
```json
{ "completed": true }
```

---

### DELETE /api/articles
Delete article (enrolled users only).  
Fails with 409 if any user has completed it.

Request:
```json
{ "articleId": 1 }
```

Response: 204 No Content

---

## ASSIGNMENTS — Auth Required

### POST /api/assignments
Create assignment (membership in owning university required).

Request:
```json
{ "courseId": 123, "title": "Assignment 1", "description": "Optional text" }
```

Response (201 Created):
```json
{ "id": 1, "courseId": 123, "title": "Assignment 1" }
```

---

### GET /api/assignments
List assignments for a course (enrolled users only).

Query: `?courseId=123`

Response (200 OK):
```json
[
  { "id": 1, "title": "Assignment 1", "completed": false }
]
```

---

### PATCH /api/assignments/{id}/deadline
Set or clear assignment deadline.

Request:
```json
{ "deadline": 1735689600 }
```

Response (200 OK):
```json
{ "id": 1, "deadline": 1735689600 }
```

---

### PATCH /api/assignments/{id}/progress
Mark assignment completion.

Request:
```json
{ "completed": true }
```

Response (200 OK):
```json
{ "completed": true }
```

---

### DELETE /api/assignments
Delete assignment (enrolled users only).  
Fails with 409 if any user has completed it.

Request:
```json
{ "assignmentId": 1 }
```

Response: 204 No Content
