import auth from "./services/auth.js";

export function mountAuth({ onAuthed }) {
  const root = document.createElement("div");
  root.className = "auth-wrap";

  const card = document.createElement("div");
  card.className = "auth-card";

  const title = document.createElement("h2");
  title.className = "auth-title";
  title.textContent = "Sign in";

  const msg = document.createElement("div");
  msg.className = "auth-message";
  msg.setAttribute("role", "alert");
  msg.setAttribute("aria-live", "polite");

  const form = document.createElement("form");
  form.className = "auth-form";

  const email = inputRow("Email", "email", "email");
  const pass = inputRow("Password", "password", "password");

  const actions = document.createElement("div");
  actions.className = "auth-actions";

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.className = "bf-btn bf-btn-primary";
  submit.textContent = "Log in";

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.className = "bf-btn";
  toggle.textContent = "Need an account? Register";

  actions.append(submit, toggle);
  form.append(email.row, pass.row, actions);

  let mode = "login";
  toggle.addEventListener("click", () => {
    mode = mode === "login" ? "register" : "login";
    title.textContent = mode === "login" ? "Sign in" : "Create account";
    submit.textContent = mode === "login" ? "Log in" : "Register";
    toggle.textContent = mode === "login" ? "Need an account? Register" : "Have an account? Log in";
    msg.textContent = "";
    msg.classList.remove("is-error");
  });

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    msg.textContent = "";
    msg.classList.remove("is-error");
    submit.disabled = true;
    try {
      const user =
        mode === "login"
          ? await auth.login(email.input.value.trim(), pass.input.value)
          : await auth.register(email.input.value.trim(), pass.input.value);
      if (typeof onAuthed === "function") onAuthed(user);
    } catch (err) {
      msg.textContent = err.message || "Something went wrong";
      msg.classList.add("is-error");
    } finally {
      submit.disabled = false;
    }
  });

  card.append(title, msg, form);
  root.appendChild(card);
  document.body.appendChild(root);

  // return unmount function
  return () => root.remove();
}

function inputRow(labelText, name, type) {
  const row = document.createElement("label");
  row.className = "bf-row";

  const label = document.createElement("span");
  label.className = "bf-label";
  label.textContent = labelText;

  const input = document.createElement("input");
  input.className = "bf-input";
  input.name = name;
  input.type = type;

  row.append(label, input);
  return { row, input };
}
