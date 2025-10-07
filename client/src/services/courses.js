class CourseService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  async getAll() {
    try {
      const res = await fetch(`${this.API_BASE}/courses`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    } catch (err) {
      console.error("Request failed:", err);
      throw err;
    }
  }

  async create({ year, term, code, name }) {
    const y = Number(year);
    const t = Number(term);
    if (!Number.isFinite(y) || y <= 0) throw new Error("Invalid year");
    if (!Number.isFinite(t) || t <= 0) throw new Error("Invalid term");
    if (!code?.trim() || !name?.trim()) throw new Error("Code and name required");

    try {
      const res = await fetch(`${this.API_BASE}/courses`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "Accept": "application/json" },
        body: JSON.stringify({ year: y, term: t, code: code.trim(), name: name.trim() }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json(); // created course
    } catch (err) {
      console.error("Create course failed:", err);
      throw err;
    }
  }

  async delete(id) {
    const cid = Number(id);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid id");

    try {
      const res = await fetch(`${this.API_BASE}/courses/${cid}`, {
        method: "DELETE",
        headers: { "Accept": "application/json" },
      });

      if (res.status === 204) return true;       // deleted
      if (res.status === 404) throw new Error("Course not found");
      throw new Error(`HTTP ${res.status}`);
    } catch (err) {
      console.error("Delete course failed:", err);
      throw err;
    }
  }
}

const API_BASE = "/api";
export default new CourseService(API_BASE);
