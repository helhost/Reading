import auth from "../services/auth.js";
import { Toast } from "../components/Toast.js";
import Button from "../components/Button.js";

export default async function LoginPage() {
  const me = await auth.me();
  if (me) {
    Toast("warn", "Already logged in");
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
  const form = el("form", { class: "auth-form" });

  const email = inputRow("Email", "email", "email");
  const pass = inputRow("Password", "password", "password");
  const pass2 = inputRow("Confirm password", "password_confirm", "password");
  pass2.row.style.display = "none"; // only in register mode

  // sensible autocomplete defaults for login
  email.input.autocomplete = "username";
  pass.input.autocomplete = "current-password";
  pass2.input.autocomplete = "new-password";

  const actions = el("div", { class: "auth-actions" });

  // Use shared Button component
  const submit = Button({ label: "Log in", type: "primary" });
  submit.type = "submit";

  const toggle = Button({ label: "Need an account? Register", type: "default" });
  toggle.type = "button";

  actions.append(submit, toggle);
  form.append(email.row, pass.row, pass2.row, actions);
  card.append(title, form);
  wrap.append(card);
  root.append(wrap);

  let mode = "login";

  toggle.addEventListener("click", () => {
    mode = mode === "login" ? "register" : "login";

    title.textContent = mode === "login" ? "Sign in" : "Create account";
    submit.textContent = mode === "login" ? "Log in" : "Register";
    toggle.textContent = mode === "login" ? "Need an account? Register"
      : "Have an account? Log in";

    // show/hide confirm password
    pass2.row.style.display = mode === "register" ? "" : "none";

    // switch autocomplete hints
    pass.input.autocomplete = mode === "register" ? "new-password" : "current-password";
    pass2.input.autocomplete = "new-password";

    // clear password fields on mode change
    pass.input.value = "";
    pass2.input.value = "";
    pass.input.focus();
  });

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    submit.disabled = true;

    const emailVal = email.input.value.trim();
    const passVal = pass.input.value;
    const pass2Val = pass2.input.value;

    if (mode === "register") {
      if (!emailVal || !passVal || !pass2Val) {
        Toast("error", "Please fill out all fields");
        submit.disabled = false;
        return;
      }
      if (passVal !== pass2Val) {
        Toast("error", "Passwords do not match");
        submit.disabled = false;
        pass2.input.focus();
        return;
      }
      if (passVal.length < 8) {
        Toast("warn", "Consider using at least 8 characters for your password");
      }
    }

    try {
      const user = mode === "login"
        ? await auth.login(emailVal, passVal)
        : await auth.register(emailVal, passVal);

      Toast("success", mode === "login" ? "Welcome back!" : "Account created!");
      window.dispatchEvent(new CustomEvent("auth:changed", { detail: { user } }));

      if (window.router?.navigate) {
        window.router.navigate("/", { callHandler: true, updateBrowserURL: true });
      } else {
        location.hash = "#/";
      }

      return user;
    } catch (err) {
      Toast("error", err?.message || "Something went wrong");
    } finally {
      submit.disabled = false;
    }
  });

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
