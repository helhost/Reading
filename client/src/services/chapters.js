class ChaptersService {
  constructor(apiBase = "/api") {
    this.API_BASE = apiBase;
  }

  /**
   * Set or clear a chapter deadline.
   * PATCH /api/chapters/{id}/deadline
   * @param {number} chapterId
   * @param {number|null} deadlineUnixSeconds - unix seconds, or null to clear
   * @returns {Promise<{id:number, deadline:number|null}>}
   */
  async setDeadline(chapterId, deadlineUnixSeconds) {
    const id = Number(chapterId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid chapterId");

    const payload =
      deadlineUnixSeconds === null
        ? { deadline: null }
        : { deadline: Number(deadlineUnixSeconds) };

    if (payload.deadline !== null && !Number.isFinite(payload.deadline)) {
      throw new Error("deadline must be a unix timestamp in seconds or null");
    }

    try {
      const res = await fetch(`${this.API_BASE}/chapters/${encodeURIComponent(String(id))}/deadline`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        if (res.status === 401) throw new Error("Unauthorized");
        if (res.status === 403) throw new Error("Forbidden");
        if (res.status === 404) throw new Error("Chapter not found");
        if (res.status === 400) throw new Error("Bad request");
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.json();
    } catch (err) {
      console.error("Chapters setDeadline failed:", err);
      throw err;
    }
  }

  /**
   * Mark a chapter as completed/incomplete.
   * PATCH /api/chapters/{id}/progress
   * @param {number} chapterId
   * @param {boolean} completed
   * @returns {Promise<{completed:boolean}>}
   */
  async setProgress(chapterId, completed) {
    const id = Number(chapterId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid chapterId");
    if (typeof completed !== "boolean") throw new Error("completed must be boolean");

    try {
      const res = await fetch(`${this.API_BASE}/chapters/${encodeURIComponent(String(id))}/progress`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ completed }),
      });

      if (!res.ok) {
        if (res.status === 401) throw new Error("Unauthorized");
        if (res.status === 403) throw new Error("Forbidden");
        if (res.status === 404) throw new Error("Chapter not found");
        if (res.status === 400) throw new Error("Bad request");
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.json();
    } catch (err) {
      console.error("Chapters setProgress failed:", err);
      throw err;
    }
  }

  // Optional convenience helpers:

  /**
   * Clear a chapter deadline.
   */
  async clearDeadline(chapterId) {
    return this.setDeadline(chapterId, null);
  }

  /**
   * Toggle progress based on current value (caller must pass current).
   */
  async toggleProgress(chapterId, currentCompleted) {
    return this.setProgress(chapterId, !currentCompleted);
  }
}

const API_BASE = "/api";
export default new ChaptersService(API_BASE);
