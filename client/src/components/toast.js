export function Toast(type, message, opts = {}) {
  const { timeout = 2000 } = opts;

  // 1) Ensure a single global container outside #app
  let container = document.getElementById("toast-root");
  if (!container) {
    container = document.createElement("div");
    container.id = "toast-root";
    container.setAttribute("aria-live", "polite");
    container.setAttribute("aria-atomic", "true");
    document.body.appendChild(container); // <-- not inside #app
  }

  // 2) Create the toast
  const el = document.createElement("div");
  el.className = `toast toast--${cssType(type)}`;
  el.role = "status";
  el.textContent = String(message);

  // 3) Dismiss handlers
  const close = () => {
    el.classList.add("toast--hide");
    el.addEventListener("transitionend", () => el.remove(), { once: true });
  };
  if (timeout > 0) setTimeout(close, timeout);

  // Click to dismiss
  el.addEventListener("click", close);

  container.appendChild(el);
  // force reflow so CSS transition can run on next frame (optional)
  // void el.offsetWidth;
  requestAnimationFrame(() => el.classList.add("toast--show"));
}

function cssType(type) {
  switch ((type || "").toLowerCase()) {
    case "success": case "ok": return "success";
    case "warn": case "warning": return "warn";
    case "error": case "danger": return "error";
    default: return "info";
  }
}
