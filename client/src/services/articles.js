class ArticlesService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  /* GET /api/articles?courseId=123 */
  async listByCourse(courseId) {
    const cid = Number(courseId);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");

    const url = `${this.API_BASE}/articles?courseId=${encodeURIComponent(cid)}`;
    const res = await fetch(url, { headers: { Accept: "application/json" } });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // [{ id, title, completed, ... }]
  }

  /* POST /api/articles */
  async create({ courseId, title, author, location }) {
    const cid = Number(courseId);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");
    if (!title?.trim()) throw new Error("Title is required");
    if (!author?.trim()) throw new Error("Author is required");

    const payload = {
      courseId: cid,
      title: title.trim(),
      author: author.trim(),
      ...(location?.trim() ? { location: location.trim() } : {}),
    };

    const res = await fetch(`${this.API_BASE}/articles`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(payload),
    });

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { id, courseId, title, author }
  }

  /* PATCH /api/articles/{id}/deadline */
  async setDeadline(articleId, deadlineOrNull) {
    const id = Number(articleId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid articleId");

    const body = { deadline: deadlineOrNull == null ? null : Number(deadlineOrNull) };
    if (body.deadline !== null && !Number.isFinite(body.deadline)) {
      throw new Error("Invalid deadline");
    }

    const res = await fetch(
      `${this.API_BASE}/articles/${encodeURIComponent(id)}/deadline`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify(body),
      }
    );

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { id, deadline }
  }

  /* PATCH /api/articles/{id}/progress */
  async setProgress(articleId, completed) {
    const id = Number(articleId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid articleId");

    const res = await fetch(
      `${this.API_BASE}/articles/${encodeURIComponent(id)}/progress`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ completed: !!completed }),
      }
    );

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { completed: true|false }
  }

  /* DELETE /api/articles  body: { articleId } */
  async delete(articleId) {
    const id = Number(articleId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid articleId");

    const res = await fetch(`${this.API_BASE}/articles`, {
      method: "DELETE",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ articleId: id }),
    });

    if (res.status === 204) return true;
    if (res.status === 404) throw new Error("Article not found");
    if (res.status === 409) throw new Error("409 Conflict");
    throw new Error(`HTTP ${res.status}`);
  }
}

const API_BASE = "/api";
export default new ArticlesService(API_BASE);
