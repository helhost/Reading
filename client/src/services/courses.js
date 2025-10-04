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
}

const API_BASE = "/api";

const courseService = new CourseService(API_BASE);
export default courseService;
