export function openModal({
  title = "",
  anchorEl = null,
  placement = "bottom-start",
  offset = 8,
  cardClass = "",
  overlayClass = "",
  onClose,
  centered = false, // add classes instead of inline styles
} = {}) {
  const cleanupFns = [];
  let closed = false;

  function close() {
    if (closed) return;
    closed = true;
    cleanupFns.forEach((fn) => { try { fn(); } catch { } });
    overlay.remove();
    if (typeof onClose === "function") onClose();
  }

  // overlay
  const overlay = document.createElement("div");
  overlay.className = [
    "modal-overlay",
    anchorEl ? "modal-overlay--clear" : "",
    (!anchorEl && centered) ? "modal-overlay--centered" : "",
    overlayClass,
  ].filter(Boolean).join(" ");
  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) close(); // click outside
  });

  // card
  const card = document.createElement("div");
  card.className = [
    "modal-card",
    anchorEl ? "modal-card--popover" : "",
    (!anchorEl && centered) ? "modal-card--centered" : "",
    cardClass,
  ].filter(Boolean).join(" ");
  card.addEventListener("click", (e) => e.stopPropagation());

  // header
  if (title && !anchorEl) {
    const header = document.createElement("div");
    header.className = "modal-header";

    const h = document.createElement("h3");
    h.className = "modal-title";
    h.textContent = title;

    const x = document.createElement("button");
    x.type = "button";
    x.className = "modal-close";
    x.textContent = "×";
    x.addEventListener("click", close);

    header.append(h, x);
    card.appendChild(header);
  } else if (anchorEl) {
    const header = document.createElement("div");
    header.className = "modal-header modal-header--compact";

    const spacer = document.createElement("div");
    const x = document.createElement("button");
    x.type = "button";
    x.className = "modal-close";
    x.textContent = "×";
    x.addEventListener("click", close);

    header.append(spacer, x);
    card.appendChild(header);
  }

  // body
  const body = document.createElement("div");
  body.className = "modal-body";
  card.appendChild(body);

  // mount
  overlay.append(card);
  document.body.append(overlay);

  // ESC to close
  const onKey = (e) => { if (e.key === "Escape") close(); };
  document.addEventListener("keydown", onKey);
  cleanupFns.push(() => document.removeEventListener("keydown", onKey));

  // anchored popover positioning
  function updatePosition() {
    if (!anchorEl) return;
    const rect = anchorEl.getBoundingClientRect();

    card.style.visibility = "hidden";
    card.style.position = "fixed";
    card.style.left = "0";
    card.style.top = "0";
    card.style.maxWidth = "min(420px, 96vw)";
    document.body.offsetHeight; // force layout

    const cardW = card.offsetWidth;
    const cardH = card.offsetHeight;
    const vw = window.innerWidth;
    const vh = window.innerHeight;

    const placeBottom = placement.startsWith("bottom");
    const placeStart = placement.endsWith("start");

    let left = placeStart ? rect.left : rect.right - cardW;
    let top = placeBottom ? rect.bottom + offset : rect.top - cardH - offset;

    if (left + cardW > vw - 8) left = vw - cardW - 8;
    if (left < 8) left = 8;
    if (top + cardH > vh - 8) top = rect.top - cardH - offset; // flip up
    if (top < 8) top = rect.bottom + offset;                    // flip down

    card.style.left = `${left}px`;
    card.style.top = `${top}px`;
    card.style.visibility = "visible";
  }

  if (anchorEl) {
    updatePosition();
    const onScroll = () => updatePosition();
    const onResize = () => updatePosition();
    window.addEventListener("scroll", onScroll, true);
    window.addEventListener("resize", onResize);
    cleanupFns.push(() => {
      window.removeEventListener("scroll", onScroll, true);
      window.removeEventListener("resize", onResize);
    });
  }

  return {
    setBody(node) {
      body.innerHTML = "";
      if (node) body.append(node);
      if (anchorEl) updatePosition();
    },
    updatePosition,
    close,
    onCleanup(fn) { if (typeof fn === "function") cleanupFns.push(fn); },
    root: overlay,
    card,
    body,
  };
}

export default openModal;
