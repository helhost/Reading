class CourseService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  /**
   * List my enrolled courses for a university.
   * GET /api/courses?universityId=uuid
   */
  async getAll(universityId) {
    const uid = String(universityId || "").trim();
    if (!uid) throw new Error("universityId is required");

    try {
      const res = await fetch(`${this.API_BASE}/courses?universityId=${encodeURIComponent(uid)}`, {
        headers: { Accept: "application/json" },
      });
      if (!res.ok) {
        if (res.status === 401) throw new Error("Unauthorized");
        if (res.status === 403) throw new Error("Membership required");
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.json();
    } catch (err) {
      console.error("Courses getAll failed:", err);
      throw err;
    }
  }

  /**
   * List the course catalog for a university (all courses).
   * GET /api/course-catalog?universityId=uuid
   */
  async getCatalog(universityId) {
    const uid = String(universityId || "").trim();
    if (!uid) throw new Error("universityId is required");

    try {
      const res = await fetch(`${this.API_BASE}/course-catalog?universityId=${encodeURIComponent(uid)}`, {
        headers: { Accept: "application/json" },
      });
      if (!res.ok) {
        if (res.status === 401) throw new Error("Unauthorized");
        if (res.status === 403) throw new Error("Membership required");
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.json();
    } catch (err) {
      console.error("Courses getCatalog failed:", err);
      throw err;
    }
  }

  /**
   * Create a course (membership in owning university required).
   * POST /api/courses
   */
  async create({ universityId, year, term, code, name }) {
    const uid = String(universityId || "").trim();
    const y = Number(year);
    const t = Number(term);
    if (!uid) throw new Error("universityId is required");
    if (!Number.isFinite(y) || y <= 0) throw new Error("Invalid year");
    if (!Number.isFinite(t) || t <= 0) throw new Error("Invalid term");
    if (!code?.trim() || !name?.trim()) throw new Error("Code and name required");

    try {
      const res = await fetch(`${this.API_BASE}/courses`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({
          universityId: uid,
          year: y,
          term: t,
          code: code.trim(),
          name: name.trim(),
        }),
      });

      if (!res.ok) {
        if (res.status === 401) throw new Error("Unauthorized");
        if (res.status === 403) throw new Error("Membership required");
        if (res.status === 400) throw new Error("Bad request");
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.json();
    } catch (err) {
      console.error("Create course failed:", err);
      throw err;
    }
  }

  /**
   * Delete a course (only if it has no books/articles/assignments).
   * DELETE /api/courses  body: { courseId }
   */
  async delete(id) {
    const courseId = Number(id);
    if (!Number.isInteger(courseId) || courseId <= 0) throw new Error("Invalid id");

    try {
      const res = await fetch(`${this.API_BASE}/courses`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ courseId }),
      });

      if (res.status === 204) return true;
      if (res.status === 404) throw new Error("Course not found");
      if (res.status === 401) throw new Error("Unauthorized");
      if (res.status === 403) throw new Error("Membership required");
      if (res.status === 409) throw new Error("Course has dependent items");
      throw new Error(`HTTP ${res.status}`);
    } catch (err) {
      console.error("Delete course failed:", err);
      throw err;
    }
  }

  /**
   * Enroll in a course (idempotent).
   * POST /api/user-courses  body: { courseId }
   */
  async enroll(courseId) {
    const cid = Number(courseId);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");

    try {
      const res = await fetch(`${this.API_BASE}/user-courses`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ courseId: cid }),
      });

      if (res.ok) return await res.json(); // { userId, courseId } for 200/201
      if (res.status === 401) throw new Error("Unauthorized");
      if (res.status === 403) throw new Error("Membership required");
      if (res.status === 404) throw new Error("Course not found");
      throw new Error(`HTTP ${res.status}`);
    } catch (err) {
      console.error("Enroll failed:", err);
      throw err;
    }
  }

  /**
   * Unenroll from a course (idempotent).
   * DELETE /api/user-courses  body: { courseId }
   */
  async unenroll(courseId) {
    const cid = Number(courseId);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");

    try {
      const res = await fetch(`${this.API_BASE}/user-courses`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ courseId: cid }),
      });

      if (res.status === 204) return true;
      if (res.status === 401) throw new Error("Unauthorized");
      if (res.status === 403) throw new Error("Membership required");
      if (res.status === 404) throw new Error("Course not found");
      throw new Error(`HTTP ${res.status}`);
    } catch (err) {
      console.error("Unenroll failed:", err);
      throw err;
    }
  }
}

const API_BASE = "/api";
export default new CourseService(API_BASE);
