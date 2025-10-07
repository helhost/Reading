import HomePage from "./pages/HomePage.js";
import LoginPage from "./pages/LoginPage.js";
import UniversityPage from "./pages/UniversityPage.js";
import UniversityHomePage from "./pages/UniversityHomePage.js";

export const router = new Navigo("/", { hash: true });
window.router = router;

router.hooks({
  before: (done) => {
    let root = document.getElementById("app");
    if (!root) {
      root = document.createElement("main");
      root.id = "app";
      document.body.appendChild(root);
    }
    root.innerHTML = "";     // <-- clears previous view
    done();
  },
});

export function initRoutes() {
  router
    .on("/", () => HomePage())
    .on("/login", () => LoginPage())
    .on("/universities", () => UniversityPage())
    .on("/universities/:slug", ({ data }) => UniversityHomePage(data.slug))
    .notFound(() => router.navigate("/"));

  router.resolve();
  router.updatePageLinks();
}
