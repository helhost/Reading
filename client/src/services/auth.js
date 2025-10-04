class AuthService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  async me() {
    const res = await fetch(`${this.API_BASE}/me`, { headers: { Accept: "application/json" } });
    if (res.status === 401) return null;
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { userId, email }
  }

  async register(email, password) {
    const res = await fetch(`${this.API_BASE}/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
      const msg = res.status === 409 ? "Email already registered" : `HTTP ${res.status}`;
      throw new Error(msg);
    }
    return this.me();
  }

  async login(email, password) {
    const res = await fetch(`${this.API_BASE}/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (res.status === 401) throw new Error("Invalid credentials");
    if (!res.ok) throw new Error(`HTTP ${res.status}`);

    return this.me();
  }

  async logout() {
    const res = await fetch(`${this.API_BASE}/logout`, {
      method: "POST",
      headers: { Accept: "application/json" },
    });
    if (!res.ok && res.status !== 204) throw new Error(`HTTP ${res.status}`);
    return true;
  }
}

const API_BASE = "/api";
export default new AuthService(API_BASE);
