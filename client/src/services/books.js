class BookService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  async getAll() {
    try {
      const res = await fetch(`${this.API_BASE}/books`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    } catch (err) {
      console.error("Request failed:", err);
      throw err;
    }
  }

  async addProgress(id, chapter) {
    try {
      const res = await fetch(`${this.API_BASE}/books/${id}/progress`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "Accept": "application/json" },
        body: JSON.stringify({ chapter, action: "add" }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json(); // returns []int of completed chapters
    } catch (err) {
      console.error("Add progress failed:", err);
      throw err;
    }
  }

  async removeProgress(id, chapter) {
    try {
      const res = await fetch(`${this.API_BASE}/books/${id}/progress`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "Accept": "application/json" },
        body: JSON.stringify({ chapter, action: "remove" }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json(); // returns []int of completed chapters
    } catch (err) {
      console.error("Remove progress failed:", err);
      throw err;
    }
  }

  async create({ courseId, title, author, numChapters, link }) {
    try {
      const res = await fetch(`${this.API_BASE}/books`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Accept": "application/json",
        },
        body: JSON.stringify({ courseId, title, author, numChapters, link }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (err) {
      console.error("Create failed:", err);
      throw err;
    }
  }

  async delete(id) {
    try {
      const res = await fetch(`${this.API_BASE}/books/${id}`, {
        method: "DELETE",
        headers: { "Accept": "application/json" },
      });
      if (!res.ok && res.status !== 204) throw new Error(`HTTP ${res.status}`);
      return true;
    } catch (err) {
      console.error("Delete failed:", err);
      throw err;
    }
  }
}

const API_BASE = "/api";
const bookService = new BookService(API_BASE);
export default bookService;
