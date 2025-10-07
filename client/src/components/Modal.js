export function openModal({ title = "", onClose } = {}) {
  // close helper
  function close() {
    cleanup();
    overlay.remove();
    if (typeof onClose === "function") onClose();
  }

  // overlay
  const overlay = document.createElement("div");
  overlay.className = "modal-overlay";
  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) close(); // click outside
  });

  // card
  const card = document.createElement("div");
  card.className = "modal-card";
  card.addEventListener("click", (e) => e.stopPropagation());

  // header
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

  // body (you’ll inject content later)
  const body = document.createElement("div");
  body.className = "modal-body";

  // assemble
  card.append(header, body);
  overlay.append(card);
  document.body.append(overlay);

  // esc to close
  function onKey(e) {
    if (e.key === "Escape") close();
  }
  document.addEventListener("keydown", onKey);

  function cleanup() {
    document.removeEventListener("keydown", onKey);
  }

  // API: let caller inject/replace content & programmatically close
  return {
    setBody(node) {
      body.innerHTML = "";
      if (node) body.append(node);
    },
    close,
    root: overlay,
    card,
    body,
  };
}

export default openModal;
