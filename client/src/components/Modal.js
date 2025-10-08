export function openModal({
  title = "",
  anchorEl = null,
  placement = "bottom-start",
  offset = 8,
  cardClass = "",
  overlayClass = "",
  onClose,
} = {}) {
  // close helper
  const cleanupFns = [];
  let closed = false;

  function close() {
    if (closed) return;
    closed = true;
    cleanupFns.forEach((fn) => {
      try { fn(); } catch { }
    });
    overlay.remove();
    if (typeof onClose === "function") onClose();
  }

  // overlay
  const overlay = document.createElement("div");
  overlay.className = `modal-overlay ${anchorEl ? "modal-overlay--clear" : ""} ${overlayClass}`.trim();
  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) close(); // click outside
  });

  // card
  const card = document.createElement("div");
  card.className = `modal-card ${anchorEl ? "modal-card--popover" : ""} ${cardClass}`.trim();
  card.addEventListener("click", (e) => e.stopPropagation());

  // header (optional)
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
    // compact header for popovers (close button only)
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

  // assemble
  overlay.append(card);
  document.body.append(overlay);

  // ESC to close
  const onKey = (e) => { if (e.key === "Escape") close(); };
  document.addEventListener("keydown", onKey);
  cleanupFns.push(() => document.removeEventListener("keydown", onKey));

  // Positioning (for popover)
  function updatePosition() {
    if (!anchorEl) return;
    const rect = anchorEl.getBoundingClientRect();

    // Make card measurable before positioning
    card.style.visibility = "hidden";
    card.style.position = "absolute";
    card.style.left = "0";
    card.style.top = "0";
    card.style.maxWidth = "min(420px, 96vw)";
    card.style.visibility = "hidden";
    // force layout
    document.body.offsetHeight; // eslint-disable-line no-unused-expressions

    const cardW = card.offsetWidth;
    const cardH = card.offsetHeight;

    let left, top;

    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const pageX = window.scrollX;
    const pageY = window.scrollY;

    const placeBottom = placement.startsWith("bottom");
    const placeStart = placement.endsWith("start");

    // initial coords
    left = pageX + (placeStart ? rect.left : rect.right - cardW);
    top = pageY + (placeBottom ? rect.bottom + offset : rect.top - cardH - offset);

    // viewport adjust (basic)
    if (left + cardW > pageX + vw - 8) left = pageX + vw - cardW - 8;
    if (left < pageX + 8) left = pageX + 8;
    if (top + cardH > pageY + vh - 8) {
      // flip to top if bottom doesn’t fit
      top = pageY + rect.top - cardH - offset;
    }
    if (top < pageY + 8) {
      // flip to bottom if top doesn’t fit either
      top = pageY + rect.bottom + offset;
    }

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

  // public API
  return {
    setBody(node) {
      body.innerHTML = "";
      if (node) body.append(node);
      if (anchorEl) updatePosition();
    },
    updatePosition,
    close,
    root: overlay,
    card,
    body,
  };
}

export default openModal;
