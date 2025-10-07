import { router, initRoutes } from "./router.js";
import auth from "./services/auth.js";
import { mountBanner, updateBanner } from "./components/Banner.js";
import { Toast } from "./components/Toast.js";

window.router = router; // helpful for programmatic nav from pages/components

async function boot() {
  const user = await safeMe();

  const links = [];

  // mount banner
  mountBanner({
    user,
    links,
    onLogout: async () => {
      try {
        await auth.logout();
        updateBanner({ user: null, links });
        Toast("success", "Logged out");
        router.navigate("/login", { callHandler: true, updateBrowserURL: true });
      } catch (e) {
        Toast("error", e?.message || "Failed to log out");
      }
      // rebind links for Navigo whenever banner DOM changes
      router.updatePageLinks();
    },
  });

  // ensure banner links are wired to Navigo
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
