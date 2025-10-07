import auth from "../services/auth.js";

export default async function LoginPage() {

  const me = await auth.me()
  if (me) {
    if (window.router?.navigate) {
      window.router.navigate("/", { callHandler: true, updateBrowserURL: true });
    } else {
      location.hash = "#/";
    }
    return;
  }

  const root = ensureRoot();
  root.innerHTML = "";

  const wrap = el("div", { class: "auth-wrap" });
  const card = el("div", { class: "auth-card" });

  const title = el("h2", { class: "auth-title", text: "Sign in" });
  const msg = el("div", { class: "auth-message" });
  msg.setAttribute("role", "alert");
  msg.setAttribute("aria-live", "polite");

  const form = el("form", { class: "auth-form" });

  const email = inputRow("Email", "email", "email");
  const pass = inputRow("Password", "password", "password");

  const actions = el("div", { class: "auth-actions" });
  const submit = el("button", { class: "bf-btn bf-btn-primary", text: "Log in" });
  submit.type = "submit";

  const toggle = el("button", { class: "bf-btn", text: "Need an account? Register" });
  toggle.type = "button";

  actions.append(submit, toggle);
  form.append(email.row, pass.row, actions);
  card.append(title, msg, form);
  wrap.append(card);
  root.append(wrap);

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

    const emailVal = email.input.value.trim();
    const passVal = pass.input.value;

    try {
      const user =
        mode === "login"
          ? await auth.login(emailVal, passVal)
          : await auth.register(emailVal, passVal);

      if (window.router?.navigate) {
        window.router.navigate("/", { callHandler: true, updateBrowserURL: true });
      } else {
        location.hash = "#/";
      }
      return user;
    } catch (err) {
      msg.textContent = err?.message || "Something went wrong";
      msg.classList.add("is-error");
    } finally {
      submit.disabled = false;
    }
  });

  // Autofocus email
  email.input.focus();
}

/* ---------- tiny DOM helpers ---------- */
function ensureRoot() {
  let node = document.getElementById("app");
  if (!node) {
    node = document.createElement("main");
    node.id = "app";
    document.body.appendChild(node);
  }
  return node;
}

function el(tag, opts = {}) {
  const n = document.createElement(tag);
  if (opts.class) n.className = opts.class;
  if (opts.text != null) n.textContent = opts.text;
  return n;
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
