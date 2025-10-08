import Button from "./Button.js";

let _openOverlay = null;

export function OpenSearchListModal({
  title = "Select",
  items = [],
  getTitle = (x) => String(x ?? ""),
  actionLabel = "Select",
  onPick = () => { },
  footerLabel,
  onFooterClick,
}) {
  closeModal(); // ensure single instance

  const overlay = document.createElement("div");
  overlay.className = "modal-overlay";
  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) closeModal();
  });

  const card = document.createElement("div");
  card.className = "modal-card";
  card.addEventListener("click", (e) => e.stopPropagation());

  // header
  const header = document.createElement("div");
  header.className = "modal-header";

  const h = document.createElement("h3");
  h.className = "modal-title";
  h.textContent = title;

  const closeBtn = document.createElement("button");
  closeBtn.className = "modal-close";
  closeBtn.type = "button";
  closeBtn.textContent = "×";
  closeBtn.addEventListener("click", closeModal);

  header.append(h, closeBtn);

  // body
  const body = document.createElement("div");
  body.className = "modal-body";

  const input = document.createElement("input");
  input.className = "searchlist__input";
  input.placeholder = "Search…";

  const list = document.createElement("div");
  list.className = "searchlist__list";

  function render(filter = "") {
    list.innerHTML = "";
    const q = filter.trim().toLowerCase();
    const filtered = q
      ? items.filter((it) => getTitle(it).toLowerCase().includes(q))
      : items;

    for (const it of filtered) {
      const row = document.createElement("div");
      row.className = "mini-card";

      const titleEl = document.createElement("div");
      titleEl.className = "mini-card__title";
      titleEl.textContent = getTitle(it);

      const btn = document.createElement("button");
      btn.className = "mini-card__btn";
      btn.type = "button";
      btn.textContent = actionLabel;
      btn.addEventListener("click", async () => {
        closeModal();
        onPick(it);
      });

      row.append(titleEl, btn);
      list.appendChild(row);
    }

    if (filtered.length === 0) {
      const msg = document.createElement("div");
      msg.className = "searchlist__empty";
      msg.textContent = "No matches found";
      msg.style.color = "var(--error)";
      list.appendChild(msg);
      return;
    }
  }

  input.addEventListener("input", () => render(input.value));
  render();

  body.append(input, list);

  // footer (OPTIONAL)
  let footer;
  if (footerLabel && typeof onFooterClick === "function") {
    footer = document.createElement("div");
    footer.className = "modal-footer";

    const footerBtn = Button({
      label: footerLabel,
      type: "primary",
      onClick: () => {
        closeModal();
        onFooterClick();
      },
    });

    footer.appendChild(footerBtn);
  }

  // assemble
  card.append(header, body, ...(footer ? [footer] : []));
  overlay.appendChild(card);
  document.body.appendChild(overlay);

  // ESC to close
  const esc = (e) => e.key === "Escape" && closeModal();
  document.addEventListener("keydown", esc);

  _openOverlay = { overlay, esc };
}

export function closeModal() {
  if (_openOverlay) {
    document.removeEventListener("keydown", _openOverlay.esc);
    _openOverlay.overlay.remove();
    _openOverlay = null;
  }
}
