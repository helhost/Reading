class BooksService {
  constructor(apiBase = "/api") {
    this.API_BASE = apiBase;
  }

  /**
   * List books for a course (enrolled users only).
   * GET /api/books?courseId=123
   */
  async getByCourse(courseId) {
    const cid = Number(courseId);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");

    try {
      const res = await fetch(
        `${this.API_BASE}/books?courseId=${encodeURIComponent(String(cid))}`,
        { headers: { Accept: "application/json" } }
      );

      if (!res.ok) {
        if (res.status === 401) throw new Error("Unauthorized");
        if (res.status === 403) throw new Error("Enrollment required");
        if (res.status === 404) throw new Error("Course not found");
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.json(); // Array<Book>
    } catch (err) {
      console.error("Books getByCourse failed:", err);
      throw err;
    }
  }

  /**
   * Create a book (membership in owning university required).
   * POST /api/books
   */
  async create({ courseId, title, author, numChapters, location }) {
    const cid = Number(courseId);
    const n = Number(numChapters);
    const t = String(title ?? "").trim();
    const a = String(author ?? "").trim();
    const loc = location == null ? undefined : String(location).trim();

    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");
    if (!t) throw new Error("Title is required");
    if (!a) throw new Error("Author is required");
    if (!Number.isInteger(n) || n <= 0) throw new Error("numChapters must be a positive integer");

    const body = {
      courseId: cid,
      title: t,
      author: a,
      numChapters: n,
      ...(loc ? { location: loc } : {}),
    };

    try {
      const res = await fetch(`${this.API_BASE}/books`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        if (res.status === 401) throw new Error("Unauthorized");
        if (res.status === 403) throw new Error("Membership required");
        if (res.status === 400) throw new Error("Bad request");
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.json(); // created { id, courseId, title, author }
    } catch (err) {
      console.error("Books create failed:", err);
      throw err;
    }
  }

  /**
   * Delete a book (enrolled users only).
   * DELETE /api/books  body: { bookId }
   * Fails with 409 if any chapter has progress.
   */
  async delete(bookId) {
    const bid = Number(bookId);
    if (!Number.isInteger(bid) || bid <= 0) throw new Error("Invalid bookId");

    try {
      const res = await fetch(`${this.API_BASE}/books`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ bookId: bid }),
      });

      if (res.status === 204) return true;
      if (res.status === 404) throw new Error("Book not found");
      if (res.status === 401) throw new Error("Unauthorized");
      if (res.status === 403) throw new Error("Enrollment required");
      if (res.status === 409) throw new Error("Cannot delete: chapter progress exists");
      throw new Error(`HTTP ${res.status}`);
    } catch (err) {
      console.error("Books delete failed:", err);
      throw err;
    }
  }
}

const API_BASE = "/api";
export default new BooksService(API_BASE);
