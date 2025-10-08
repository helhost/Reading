import { router, initRoutes } from "./router.js";
import auth from "./services/auth.js";
import { mountBanner, updateBanner } from "./components/Banner.js";
import { Toast } from "./components/Toast.js";

window.router = router; // helpful for programmatic nav from pages/components

async function boot() {
  const user = await safeMe();

  // mount banner (no center links; onHomeClick is mandatory)
  mountBanner({
    user,
    onHomeClick: () => {
      router.navigate("/", { callHandler: true, updateBrowserURL: true });
    },
    onLogout: async () => {
      try {
        await auth.logout();
        updateBanner({ user: null });
        Toast("success", "Logged out");
        router.navigate("/login", { callHandler: true, updateBrowserURL: true });
      } catch (e) {
        Toast("error", e?.message || "Failed to log out");
      }
      // banner DOM changed (buttons re-created), so rebind Navigo
      router.updatePageLinks();
    },
  });

  // keep banner in sync when auth changes anywhere
  window.addEventListener("auth:changed", async (e) => {
    const nextUser = e?.detail?.user ?? (await safeMe());
    updateBanner({ user: nextUser });
    // banner DOM may have changed (links/buttons), rebind Navigo
    router.updatePageLinks();
  });

  // ensure any other page links are wired to Navigo
  router.updatePageLinks();

  // start routes
  initRoutes();
}

async function safeMe() {
  try {
    return await auth.me(); // { userId, email } or null
  } catch {
    return null;
  }
}

boot();
