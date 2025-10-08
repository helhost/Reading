import { router, initRoutes } from "./router.js";
import auth from "./services/auth.js";
import { mountBanner, updateBanner } from "./components/Banner.js";
import { Toast } from "./components/Toast.js";
import { mountCalendarSubscribe, unmountCalendarSubscribe } from "./components/CalendarSubscribe.js";

window.router = router;

async function boot() {
  const user = await safeMe();

  mountBanner({
    user,
    onHomeClick: () => {
      router.navigate("/", { callHandler: true, updateBrowserURL: true });
    },
    onLogout: async () => {
      try {
        await auth.logout();
        updateBanner({ user: null });
        // notify app so subscribers can react
        window.dispatchEvent(new CustomEvent("auth:changed", { detail: { user: null } }));
        Toast("success", "Logged out");
        router.navigate("/login", { callHandler: true, updateBrowserURL: true });
      } catch (e) {
        Toast("error", e?.message || "Failed to log out");
      }
      router.updatePageLinks();
    },
  });

  // ensure correct initial state
  if (user) mountCalendarSubscribe(); else unmountCalendarSubscribe();

  // react to login/register/logout everywhere
  window.addEventListener("auth:changed", async (e) => {
    const nextUser = e?.detail?.user ?? (await safeMe());
    updateBanner({ user: nextUser });

    if (nextUser) mountCalendarSubscribe(); else unmountCalendarSubscribe();

    router.updatePageLinks();
  });

  router.updatePageLinks();
  initRoutes();
}

async function safeMe() {
  try { return await auth.me(); } catch { return null; }
}

boot();
