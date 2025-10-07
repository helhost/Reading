let _escHandler = null;

export function openBookForm(courseId, opts = {}) {
  const {
    onSubmit,
    title = "Add book",
    initial = {}, // { title, author, numChapters, link }
  } = opts;

  closeBookForm(); // replace if already open

  // overlay
  const overlay = document.createElement("div");
  overlay.className = "bookform-overlay";
  overlay.addEventListener("click", closeBookForm);

  // card
  const card = document.createElement("div");
  card.className = "bookform-card";
  card.addEventListener("click", (e) => e.stopPropagation());

  // header
  const header = document.createElement("div");
  header.className = "bookform-header";

  const titleEl = document.createElement("h3");
  titleEl.className = "bookform-title";
  titleEl.textContent = title;

  const closeBtn = document.createElement("button");
  closeBtn.type = "button";
  closeBtn.className = "bookform-close";
  closeBtn.textContent = "×";
  closeBtn.addEventListener("click", closeBookForm);

  header.append(titleEl, closeBtn);

  // body + form
  const body = document.createElement("div");
  body.className = "bookform-body";

  const form = document.createElement("form");
  form.className = "book-form";

  form.append(
    makeField("Title", "title", "text", initial.title || "", { required: true }),
    makeField("Author", "author", "text", initial.author || "", { required: true }),
    makeField(
      "Number of chapters",
      "numChapters",
      "number",
      initial.numChapters ?? "",
      { required: true, min: 1, step: 1 }
    ),
    //makeField("Link (optional)", "link", "url", initial.link || "")
  );

  // actions INSIDE the form (native submit)
  const actionsEl = document.createElement("div");
  actionsEl.className = "bookform-actions";

  const cancel = document.createElement("button");
  cancel.type = "button";
  cancel.className = "bf-btn";
  cancel.textContent = "Cancel";
  cancel.addEventListener("click", closeBookForm);

  const create = document.createElement("button");
  create.type = "submit";
  create.className = "bf-btn bf-btn-primary";
  create.textContent = "Create";

  actionsEl.append(cancel, create);
  form.appendChild(actionsEl);

  // submit (collect values via FormData)
  form.addEventListener("submit", (e) => {
    e.preventDefault();
    const fd = new FormData(form);
    const data = {
      courseId,
      title: (fd.get("title") || "").toString().trim(),
      author: (fd.get("author") || "").toString().trim(),
      numChapters: (fd.get("numChapters") || "").toString().trim(),
      //link: (fd.get("link") || "").toString().trim(),
    };
    if (typeof onSubmit === "function") onSubmit(data, { close: closeBookForm });
    else { console.log("create book (preview):", data); closeBookForm(); }
  });

  // disable Create until valid
  function setState() { create.disabled = !form.checkValidity(); }
  form.addEventListener("input", setState);
  setState();

  // assemble
  body.appendChild(form);
  card.append(header, body);
  overlay.appendChild(card);
  document.body.appendChild(overlay);
  document.body.classList.add("bookform-open");

  // ESC to close
  _escHandler = (e) => { if (e.key === "Escape") closeBookForm(); };
  document.addEventListener("keydown", _escHandler);

  // focus first field
  const firstInput = form.querySelector("input[name='title']");
  if (firstInput) firstInput.focus();
}

export function closeBookForm() {
  const overlay = document.querySelector(".bookform-overlay");
  if (overlay) overlay.remove();
  document.body.classList.remove("bookform-open");
  if (_escHandler) {
    document.removeEventListener("keydown", _escHandler);
    _escHandler = null;
  }
}

// internal
function makeField(labelText, name, type, value = "", attrs = {}) {
  const row = document.createElement("label");
  row.className = "bf-row";

  const label = document.createElement("span");
  label.className = "bf-label";
  label.textContent = labelText;

  const input = document.createElement("input");
  input.className = "bf-input";
  input.name = name;
  input.type = type;

  if (value !== undefined && value !== null && value !== "") input.value = value;
  for (const [k, v] of Object.entries(attrs)) input[k] = v;

  row.append(label, input);
  return row;
}
