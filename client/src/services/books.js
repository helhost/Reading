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
}

const API_BASE = "/api";

const bookService = new BookService(API_BASE);
export default bookService;
