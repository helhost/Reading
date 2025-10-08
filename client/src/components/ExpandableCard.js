export default function ExpandableCard({
  title = "",
  subtitle = "",
  content = null,
  initiallyOpen = false,
  onToggle = null,
} = {}) {
  const root = document.createElement("section");
  // Reuse your .card look; add our own namespace class
  root.className = "card exp-card";
  root.setAttribute("aria-expanded", initiallyOpen ? "true" : "false");

  // --- Header ---
  const header = document.createElement("div");
  header.className = "exp-card__header";

  const h = document.createElement("h3");
  h.className = "card-title exp-card__title";
  h.textContent = title;

  const sub = document.createElement("div");
  sub.className = "exp-card__subtitle";
  sub.textContent = subtitle;

  // top-right indicator (▾ rotates when collapsed)
  const indicator = document.createElement("span");
  indicator.className = "exp-card__indicator";
  indicator.textContent = "▾";
  indicator.setAttribute("aria-hidden", "true");

  header.append(h, sub, indicator);

  // --- Content slot ---
  const body = document.createElement("div");
  body.className = "exp-card__content";

  if (content instanceof Node) {
    body.appendChild(content);
  } else if (typeof content === "function") {
    const node = content();
    if (node) body.appendChild(node);
  }

  // --- Toggle behavior (click anywhere on the card) ---
  const setOpen = (open) => {
    root.setAttribute("aria-expanded", open ? "true" : "false");
    body.hidden = !open; // keep it simple; CSS polish comes next
    onToggle?.(open);
  };
  setOpen(!!initiallyOpen);

  root.addEventListener("click", (e) => {
    // Let inner controls opt out of toggling
    if (e.target.closest("[data-no-toggle]")) return;
    const open = root.getAttribute("aria-expanded") === "true";
    setOpen(!open);
  });

  root.append(header, body);

  // Expose a tiny imperative API (handy for parent pages)
  root.open = () => setOpen(true);
  root.close = () => setOpen(false);
  root.toggle = () => setOpen(root.getAttribute("aria-expanded") !== "true");
  root.setContent = (node) => {
    body.innerHTML = "";
    if (node) body.appendChild(node);
  };

  return root;
}
