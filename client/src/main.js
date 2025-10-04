import auth from "./services/auth.js";
import { mountAuth } from "./authView.js";

function mountUserBar(user) {
  const bar = document.createElement("div");
  bar.className = "userbar";

  const left = document.createElement("div");
  left.className = "userbar__left";
  left.textContent = user?.email ? `${user.email}` : "Logged in";

  const right = document.createElement("div");
  right.className = "userbar__right";

  const out = document.createElement("button");
  out.type = "button";
  out.className = "bf-btn";
  out.textContent = "Log out";
  out.addEventListener("click", async () => {
    await auth.logout();
    location.reload();
  });

  right.appendChild(out);
  bar.append(left, right);
  document.body.appendChild(bar);

  return () => bar.remove();
}

const user = await auth.me();
if (!user) {
  const unmountAuth = mountAuth({
    onAuthed: async () => {
      unmountAuth();
      const authed = await auth.me();
      mountUserBar(authed);
      await import("./app.js");
    },
  });
} else {
  mountUserBar(user);
  await import("./app.js");
}
