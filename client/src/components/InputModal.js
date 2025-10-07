import Button from "./Button.js";

let _open = null;

export function OpenInputModal({
  title = "Create",
  placeholder = "",
  initialValue = "",
  actionLabel = "Create",
  validate = (v) => (v.trim() ? null : "Required"),
  onSubmit = async () => { },
}) {
  closeInputModal(); // ensure single instance

  // overlay
  const overlay = document.createElement("div");
  overlay.className = "modal-overlay";
  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) closeInputModal();
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

  const closeBtn = document.createElement("button");
  closeBtn.className = "modal-close";
  closeBtn.type = "button";
  closeBtn.textContent = "×";
  closeBtn.addEventListener("click", closeInputModal);

  header.append(h, closeBtn);

  // body
  const body = document.createElement("div");
  body.className = "modal-body";

  const row = document.createElement("label");

  const input = document.createElement("input");
  input.className = "input-field";
  input.placeholder = placeholder;
  input.value = initialValue;

  const err = document.createElement("div");
  err.className = "field-error";

  row.append(input);
  body.append(row, err);

  // actions
  const actions = document.createElement("div");
  actions.className = "modal-actions";

  const cancelBtn = Button({ label: "Cancel", onClick: closeInputModal });
  const submitBtn = Button({
    label: actionLabel,
    type: "primary",
    onClick: trySubmit,
  });

  actions.append(cancelBtn, submitBtn);

  // assemble
  card.append(header, body, actions);
  overlay.appendChild(card);
  document.body.appendChild(overlay);

  const onEsc = (e) => e.key === "Escape" && closeInputModal();
  const onEnter = (e) => e.key === "Enter" && trySubmit();
  document.addEventListener("keydown", onEsc);
  input.addEventListener("keydown", onEnter);
  input.focus();

  _open = { overlay, onEsc, onEnter };

  async function trySubmit() {
    err.textContent = "";
    const val = input.value.trim();
    const vMsg = validate ? validate(val) : null;
    if (vMsg) {
      err.textContent = vMsg;
      input.focus();
      return;
    }
    try {
      submitBtn.disabled = true;
      await onSubmit(val);
      closeInputModal();
    } catch (e) {
      err.textContent = e?.message || "Something went wrong.";
    } finally {
      submitBtn.disabled = false;
    }
  }
}

export function closeInputModal() {
  if (_open) {
    document.removeEventListener("keydown", _open.onEsc);
    document.removeEventListener("keydown", _open.onEnter);
    _open.overlay.remove();
    _open = null;
  }
}
