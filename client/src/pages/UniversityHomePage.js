import getMe from "../util/index.js";
import SubBanner from "../components/SubBanner.js";
import courseService from "../services/courses.js";
import { Toast } from "../components/Toast.js";

import BookPage from "./BookPage.js";
import ArticlePage from "./ArticlePage.js";
import AssignmentPage from "./AssignmentPage.js";

export default async function UniversityHomePage(slug, tab /* "books"|"articles"|"assignments"|null */) {
  const me = await getMe();
  if (!me) return;

  // default to "books" if no tab provided
  if (!tab) {
    window.router?.navigate?.(`/universities/${slug}/books`, {
      callHandler: true,
      updateBrowserURL: true,
    });
    return;
  }

  const root = ensureRoot();
  root.innerHTML = "";

  const basePath = `/universities/${slug}`;
  const sub = SubBanner({ basePath, active: tab });

  // Simple content slot for now
  const content = document.createElement("section");
  content.className = "page-body";
  content.style.padding = "1rem";

  root.append(sub, content);

  // Dispatch to the correct stub page: Books/Articles/Assignments
  const handlers = {
    books: BookPage,
    articles: ArticlePage,
    assignments: AssignmentPage,
  };

  const handler = handlers[tab];
  if (!handler) {
    // unknown tab -> normalize to books
    window.router?.navigate?.(`/universities/${slug}/books`, {
      callHandler: true,
      updateBrowserURL: true,
    });
    return;
  }

  try {
    const { myCourses, remainingCourses } = await getCourseLists(slug);
    handler(myCourses, remainingCourses); // each stub just logs for now
  } catch (err) {
    console.error(`[${tab}] failed to load:`, err);
    Toast("error", err?.message || "Failed to load courses");
  }

  // let Navigo own the links
  window.router?.updatePageLinks?.();
}

async function getCourseLists(universityId) {
  const [myCourses, catalog] = await Promise.all([
    courseService.getAll(universityId),
    courseService.getCatalog(universityId),
  ]);
  const myIds = new Set(myCourses.map(c => c.id));
  const remainingCourses = catalog.filter(c => !myIds.has(c.id));
  return { myCourses, remainingCourses };
}

function ensureRoot() {
  let node = document.getElementById("app");
  if (!node) {
    node = document.createElement("main");
    node.id = "app";
    document.body.appendChild(node);
  }
  return node;
}
