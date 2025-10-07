import Button from "./Button.js";

let _openForm = null;

export function OpenFormModal({
  title = "Form",
  submitLabel = "Submit",
  cancelLabel = "Cancel",
  fields = [],
  onSubmit,
} = {}) {
  if (!Array.isArray(fields) || fields.length === 0) {
    throw new Error("OpenFormModal: fields[] is required");
  }
  if (typeof onSubmit !== "function") {
    throw new Error("OpenFormModal: onSubmit callback is required");
  }

  closeFormModal(); // ensure single instance

  // overlay
  const overlay = document.createElement("div");
  overlay.className = "modal-overlay";
  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) closeFormModal();
  });

  // card
  const card = document.createElement("div");
  card.className = "modal-card form-modal";
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
  closeBtn.addEventListener("click", closeFormModal);

  header.append(h, closeBtn);

  // body + form
  const body = document.createElement("div");
  body.className = "modal-body";

  const form = document.createElement("form");
  form.className = "form";
  form.setAttribute("novalidate", "novalidate");

  // build inputs
  const nameCounts = new Map();
  const rows = [];

  const toName = (label, explicit) => {
    if (explicit) return explicit;
    const base =
      String(label).toLowerCase().trim().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "") ||
      "field";
    const n = (nameCounts.get(base) || 0) + 1;
    nameCounts.set(base, n);
    return n === 1 ? base : `${base}-${n}`;
  };

  for (const f of fields) {
    const { label, type, required = false, name, placeholder, initial } = f || {};
    if (!label || (type !== "int" && type !== "string")) {
      throw new Error("OpenFormModal: each field needs { label, type: 'int'|'string' }");
    }
    const fieldName = toName(label, name);

    const row = document.createElement("label");
    row.className = "form-row";

    const lab = document.createElement("span");
    lab.className = "form-label";
    lab.textContent = label;

    const input = document.createElement("input");
    input.className = "input-field form-input";
    input.name = fieldName;

    if (type === "int") {
      input.type = "number";
      input.step = "1";
      input.inputMode = "numeric";
      input.pattern = "\\d*";
    } else {
      input.type = "text";
    }

    if (required) input.required = true;
    if (placeholder) input.placeholder = placeholder;
    if (initial !== undefined && initial !== null && initial !== "") {
      input.value = String(initial);
    }

    row.append(lab, input);
    rows.push({ cfg: f, input, fieldName });
    form.appendChild(row);
  }

  // footer
  const footer = document.createElement("div");
  footer.className = "modal-footer form-footer";

  const cancel = Button({
    label: cancelLabel,
    type: "default",
    onClick: () => closeFormModal(),
  });

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "btn btn--primary";
  submit.textContent = submitLabel;

  footer.append(cancel, submit);

  // state + helpers
  let submitting = false;

  const setSubmitting = (is) => {
    submitting = is;
    submit.disabled = is || submit.disabled; // keep disabled if invalid
    cancel.disabled = is;
    rows.forEach(({ input }) => (input.disabled = is));
  };

  const updateSubmitState = () => {
    if (submitting) return;
    const allGood = rows.every(({ cfg, input }) => {
      if (!cfg.required) return true;
      const v = input.value.trim();
      if (cfg.type === "int") {
        const n = Number.parseInt(v, 10);
        return v !== "" && Number.isFinite(n);
      }
      return v.length > 0;
    });
    submit.disabled = !allGood;
  };

  form.addEventListener("input", updateSubmitState);
  // initial check
  updateSubmitState();

  // submit handler
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (submitting) return;

    // field-level validity
    for (const { cfg, input } of rows) {
      input.setCustomValidity("");
      if (cfg.required && !input.value.trim()) {
        input.setCustomValidity("Required");
      }
      if (cfg.type === "int" && input.value.trim() !== "") {
        const n = Number.parseInt(input.value.trim(), 10);
        if (!Number.isFinite(n)) {
          input.setCustomValidity("Enter a whole number");
        }
      }
      if (!input.checkValidity()) {
        input.reportValidity();
        input.focus();
        return;
      }
    }

    // payload (omit optional empties)
    const data = {};
    for (const { cfg, input, fieldName } of rows) {
      const raw = input.value.trim();
      if (!raw && !cfg.required) continue;
      if (cfg.type === "int") {
        data[fieldName] = Number.parseInt(raw, 10);
      } else {
        data[fieldName] = raw;
      }
    }

    try {
      setSubmitting(true);
      await onSubmit(data, { close: closeFormModal });
    } catch (err) {
      console.error("Form submit failed:", err);
    } finally {
      setSubmitting(false);
      updateSubmitState();
    }
  });

  // assemble
  body.appendChild(form);
  card.append(header, body, footer);
  overlay.appendChild(card);
  document.body.appendChild(overlay);

  // ESC to close (but not while submitting)
  const esc = (e) => e.key === "Escape" && !submitting && closeFormModal();
  document.addEventListener("keydown", esc);

  _openForm = { overlay, esc };

  // focus first input
  const firstInput = form.querySelector("input");
  if (firstInput) firstInput.focus();
}

export function closeFormModal() {
  if (_openForm) {
    document.removeEventListener("keydown", _openForm.esc);
    _openForm.overlay.remove();
    _openForm = null;
  }
}
