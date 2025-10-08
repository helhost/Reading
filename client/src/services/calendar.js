class CalendarService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  // Get or create the user's token (auth required).
  async getToken() {
    const res = await fetch(`${this.API_BASE}/calendar/token`, {
      headers: { Accept: "application/json" },
      credentials: "include",
    });
    if (res.status === 401) return null;
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { token, urlPath }
  }

  // Rotate the token (invalidates the old public URL).
  async rotateToken() {
    const res = await fetch(`${this.API_BASE}/calendar/token/rotate`, {
      method: "POST",
      headers: { Accept: "application/json" },
      credentials: "include",
    });
    if (res.status === 401) throw new Error("Unauthorized");
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { token, urlPath }
  }

  // Build an absolute URL you can show/copy in the UI.
  toAbsoluteUrl(urlPath) {
    // Works behind reverse proxies and in dev
    return new URL(urlPath, window.location.origin).toString();
  }
}

const API_BASE = "/api";
export default new CalendarService(API_BASE);
