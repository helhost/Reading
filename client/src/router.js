import HomePage from "./pages/HomePage.js";
import LoginPage from "./pages/LoginPage.js";
import UniversityPage from "./pages/UniversityPage.js";
import UniversityHomePage from "./pages/UniversityHomePage.js";


export const router = new Navigo("/", { hash: true });

export function initRoutes() {
  router
    .on("/", () => HomePage())
    .on("/login", () => LoginPage())
    .on("/universities", () => UniversityPage())
    .on("/universities/:slug", ({ data }) => UniversityHomePage(data.slug))
    .notFound(() => console.log("route: 404"));

  router.resolve();
  router.updatePageLinks();
}
