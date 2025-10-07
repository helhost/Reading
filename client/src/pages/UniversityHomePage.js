import getMe from "../util/index.js";
import SubBanner from "../components/SubBanner.js";

export default async function UniversityHomePage(slug, tab /* "books"|"articles"|"assignments"|null */) {
  const me = await getMe();
  if (!me) return;

  const root = ensureRoot();
  root.innerHTML = "";

  const basePath = `/universities/${slug}`;
  const sub = SubBanner({ basePath, active: tab });

  // Simple content slot for now
  const content = document.createElement("section");
  content.className = "page-body";
  content.style.padding = "1rem";
  content.textContent = tab ? `Tab: ${tab}` : "Choose a section.";

  root.append(sub, content);

  // let Navigo own the links
  window.router?.updatePageLinks?.();
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
