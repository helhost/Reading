
let _escHandlerCourse = null;

export function openCourseForm(opts = {}) {
  const {
    onSubmit,
    title = "Add course",
    initial = {}, // { code, name, term, year }
  } = opts;

  closeCourseForm();

  const overlay = document.createElement("div");
  overlay.className = "bookform-overlay";
  overlay.addEventListener("click", closeCourseForm);

  const card = document.createElement("div");
  card.className = "bookform-card";
  card.addEventListener("click", (e) => e.stopPropagation());

  const header = document.createElement("div");
  header.className = "bookform-header";

  const titleEl = document.createElement("h3");
  titleEl.className = "bookform-title";
  titleEl.textContent = title;

  const closeBtn = document.createElement("button");
  closeBtn.type = "button";
  closeBtn.className = "bookform-close";
  closeBtn.textContent = "×";
  closeBtn.addEventListener("click", closeCourseForm);

  header.append(titleEl, closeBtn);

  const body = document.createElement("div");
  body.className = "bookform-body";

  const form = document.createElement("form");
  form.className = "book-form";

  form.append(
    makeField("Code", "code", "text", initial.code || "", { required: true }),
    makeField("Name", "name", "text", initial.name || "", { required: true }),
    makeField("Term", "term", "number", initial.term ?? "", { required: true, min: 1, step: 1 }),
    makeField("Year", "year", "number", initial.year ?? "", { required: true, min: 1, step: 1 })
  );

  const actionsEl = document.createElement("div");
  actionsEl.className = "bookform-actions";

  const cancel = document.createElement("button");
  cancel.type = "button";
  cancel.className = "bf-btn";
  cancel.textContent = "Cancel";
  cancel.addEventListener("click", closeCourseForm);

  const create = document.createElement("button");
  create.type = "submit";
  create.className = "bf-btn bf-btn-primary";
  create.textContent = "Create";

  actionsEl.append(cancel, create);
  form.appendChild(actionsEl);

  form.addEventListener("submit", (e) => {
    e.preventDefault();
    const fd = new FormData(form);
    const data = {
      code: (fd.get("code") || "").toString().trim(),
      name: (fd.get("name") || "").toString().trim(),
      term: (fd.get("term") || "").toString().trim(),
      year: (fd.get("year") || "").toString().trim(),
    };
    if (typeof onSubmit === "function") onSubmit(data, { close: closeCourseForm });
    else { console.log("create course (preview):", data); closeCourseForm(); }
  });

  function setState() { create.disabled = !form.checkValidity(); }
  form.addEventListener("input", setState);
  setState();

  body.appendChild(form);
  card.append(header, body);
  overlay.appendChild(card);
  document.body.appendChild(overlay);
  document.body.classList.add("bookform-open");

  _escHandlerCourse = (e) => { if (e.key === "Escape") closeCourseForm(); };
  document.addEventListener("keydown", _escHandlerCourse);

  const firstInput = form.querySelector("input[name='code']");
  if (firstInput) firstInput.focus();
}

export function closeCourseForm() {
  const overlay = document.querySelector(".bookform-overlay");
  if (overlay) overlay.remove();
  document.body.classList.remove("bookform-open");
  if (_escHandlerCourse) {
    document.removeEventListener("keydown", _escHandlerCourse);
    _escHandlerCourse = null;
  }
}

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
