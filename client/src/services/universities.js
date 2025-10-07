class UniversityService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  async getMyUniversities() {
    const res = await fetch(`${this.API_BASE}/user-universities`, {
      headers: { Accept: "application/json" },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // [{ userId, universityId, ... }]
  }

  async getAll() {
    const res = await fetch(`${this.API_BASE}/universities`, {
      headers: { Accept: "application/json" },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // [{ id, name }]
  }

  async create(name) {
    const res = await fetch(`${this.API_BASE}/universities`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ name }),
    });
    if (!res.ok) {
      // 409 is common for duplicate names; surface a clear message
      if (res.status === 409) throw new Error("University name already exists");
      throw new Error(`HTTP ${res.status}`);
    }
    return res.json(); // { id, name, created_at? }
  }

  async join(universityId) {
    const res = await fetch(`${this.API_BASE}/user-universities`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ universityId }),
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { userId, universityId, role }
  }

  async leave(universityId) {
    const res = await fetch(`${this.API_BASE}/user-universities`, {
      method: "DELETE",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ universityId }),
    });
    if (res.status === 204) return true;
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return true;
  }
}

const API_BASE = "/api";
const universityService = new UniversityService(API_BASE);
export default universityService;
