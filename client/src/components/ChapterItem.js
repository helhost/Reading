export default function ChapterItem({
  index,                // required: 1..N
  chapterId = null,     // optional backend id
  completed = false,
  deadline = null,      // unix seconds or null
  onClick,              // (ctx) => void|Promise<void>
  disabled = false,
} = {}) {
  if (!Number.isInteger(index) || index <= 0) {
    throw new Error("ChapterItem: index must be a positive integer");
  }

  const el = document.createElement("div");
  el.className = "chapter-box";
  el.textContent = String(index);
  el.setAttribute("data-no-toggle", "");   // don't collapse parent card
  el.tabIndex = 0;
  el.role = "button";
  el.ariaPressed = completed ? "true" : "false";

  if (completed) el.classList.add("completed");
  if (Number.isInteger(chapterId) && chapterId > 0) {
    el.dataset.chapterId = String(chapterId);
  }

  // init deadline state (adds .has-deadline and .deadline-past as needed)
  if (deadline != null && Number.isFinite(Number(deadline))) {
    applyDeadline(el, Number(deadline));
  } else {
    updateTitle(el);
  }

  if (disabled) el.setAttribute("aria-disabled", "true");

  // Imperative micro-API
  el.getState = () => ({
    index,
    chapterId: el.dataset.chapterId ? Number(el.dataset.chapterId) : null,
    completed: el.classList.contains("completed"),
    deadline: el.dataset.deadline ? Number(el.dataset.deadline) : null,
  });

  el.setCompleted = (v) => {
    if (v) el.classList.add("completed");
    else el.classList.remove("completed");
    el.ariaPressed = v ? "true" : "false";
    updateTitle(el);
  };

  el.toggleCompleted = () => el.setCompleted(!el.classList.contains("completed"));

  el.setDeadline = (tsOrNull) => {
    if (tsOrNull == null) {
      delete el.dataset.deadline;
      el.classList.remove("has-deadline", "deadline-past");
      updateTitle(el);
      return;
    }
    const ts = Number(tsOrNull);
    if (!Number.isFinite(ts)) return;
    applyDeadline(el, ts);
  };

  const handleActivate = (e) => {
    if (disabled) return;
    onClick?.({
      event: e,
      element: el,
      index,
      chapterId: el.dataset.chapterId ? Number(el.dataset.chapterId) : null,
      completed: el.classList.contains("completed"),
      deadline: el.dataset.deadline ? Number(el.dataset.deadline) : null,
    });
  };

  el.addEventListener("click", (e) => {
    e.stopPropagation();
    handleActivate(e);
  });

  el.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleActivate(e);
    }
  });

  return el;
}

/* ---------- helpers ---------- */

function applyDeadline(el, ts) {
  el.dataset.deadline = String(ts);
  el.classList.add("has-deadline");
  if (isPast(ts)) el.classList.add("deadline-past");
  else el.classList.remove("deadline-past");
  updateTitle(el);
}

function updateTitle(el) {
  const hasDeadline = "deadline" in el.dataset;
  const completed = el.classList.contains("completed");
  if (hasDeadline) {
    const ts = Number(el.dataset.deadline);
    el.title = `Deadline: ${formatDeadline(ts)}${completed ? " • Completed" : ""}`;
  } else {
    el.title = completed ? "Completed" : "";
  }
}

// Past means strictly before today's local midnight (in user's local time)
function isPast(unixSeconds) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const startOfToday = Math.floor(today.getTime() / 1000);
  return unixSeconds < startOfToday;
}

function formatDeadline(unixSeconds) {
  try {
    const d = new Date(unixSeconds * 1000);
    return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "2-digit" });
  } catch {
    return String(unixSeconds);
  }
}
