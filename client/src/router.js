import HomePage from "./pages/HomePage.js";
import LoginPage from "./pages/LoginPage.js";
import UniversityPage from "./pages/UniversityPage.js";
import UniversityHomePage from "./pages/UniversityHomePage.js";

export const router = new Navigo("/", { hash: true });

export function initRoutes() {
  window.router = router; // (so pages can call updatePageLinks)

  router
    .on("/", () => HomePage())
    .on("/login", () => LoginPage())
    .on("/universities", () => UniversityPage())

    // base (no tab)
    .on("/universities/:slug", ({ data }) => UniversityHomePage(data.slug, null))

    // tabbed
    .on("/universities/:slug/:tab", ({ data }) =>
      UniversityHomePage(data.slug, (data.tab || "").toLowerCase())
    )

    .notFound(() => console.log("route: 404"));

  router.resolve();
  router.updatePageLinks();
}
