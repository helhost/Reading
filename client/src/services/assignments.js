class AssignmentsService {
  constructor(apiBase) {
    this.API_BASE = apiBase;
  }

  /* GET /api/assignments?courseId=123 */
  async listByCourse(courseId) {
    const cid = Number(courseId);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");

    const url = `${this.API_BASE}/assignments?courseId=${encodeURIComponent(cid)}`;
    const res = await fetch(url, { headers: { Accept: "application/json" } });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // [{ id, title, completed, deadline? }, ...]
  }

  /* POST /api/assignments */
  async create({ courseId, title, description }) {
    const cid = Number(courseId);
    if (!Number.isInteger(cid) || cid <= 0) throw new Error("Invalid courseId");
    if (!title?.trim()) throw new Error("Title is required");

    const payload = {
      courseId: cid,
      title: title.trim(),
      ...(description?.trim() ? { description: description.trim() } : {}),
    };

    const res = await fetch(`${this.API_BASE}/assignments`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(payload),
    });

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { id, courseId, title }
  }

  /* PATCH /api/assignments/{id}/deadline */
  async setDeadline(assignmentId, deadlineOrNull) {
    const id = Number(assignmentId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid assignmentId");

    const body = { deadline: deadlineOrNull == null ? null : Number(deadlineOrNull) };
    if (body.deadline !== null && !Number.isFinite(body.deadline)) {
      throw new Error("Invalid deadline");
    }

    const res = await fetch(
      `${this.API_BASE}/assignments/${encodeURIComponent(id)}/deadline`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify(body),
      }
    );

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { id, deadline }
  }

  /* PATCH /api/assignments/{id}/progress */
  async setProgress(assignmentId, completed) {
    const id = Number(assignmentId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid assignmentId");

    const res = await fetch(
      `${this.API_BASE}/assignments/${encodeURIComponent(id)}/progress`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ completed: !!completed }),
      }
    );

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json(); // { completed: true|false }
  }

  /* DELETE /api/assignments  body: { assignmentId } */
  async delete(assignmentId) {
    const id = Number(assignmentId);
    if (!Number.isInteger(id) || id <= 0) throw new Error("Invalid assignmentId");

    const res = await fetch(`${this.API_BASE}/assignments`, {
      method: "DELETE",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ assignmentId: id }),
    });

    if (res.status === 204) return true;
    if (res.status === 404) throw new Error("Assignment not found");
    if (res.status === 409) throw new Error("409 Conflict");
    throw new Error(`HTTP ${res.status}`);
  }
}

const API_BASE = "/api";
export default new AssignmentsService(API_BASE);
