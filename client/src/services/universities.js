class UniversityService {
  constructor(apiBase = "/api") {
    this.API_BASE = apiBase;
  }

  async getMyUniversities() {
    const res = await fetch(`${this.API_BASE}/user-universities`, {
      headers: { Accept: "application/json" },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // [{ userId, universityId }]
  }

  async getAll() {
    const res = await fetch(`${this.API_BASE}/universities`, {
      headers: { Accept: "application/json" },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // [{ id, name }]
  }
}

const universityService = new UniversityService();
export default universityService;
