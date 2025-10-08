export default function SubBanner({ basePath, active = null }) {
  // normalize
  const act = (active || "").toLowerCase();

  const wrap = document.createElement("nav");
  wrap.className = "subbanner";

  wrap.innerHTML = `
    <a class="subtab ${act === "books" ? "is-active" : ""}" href="${basePath}/books" data-navigo>Books</a>
    <a class="subtab ${act === "articles" ? "is-active" : ""}" href="${basePath}/articles" data-navigo>Articles</a>
    <a class="subtab ${act === "assignments" ? "is-active" : ""}" href="${basePath}/assignments" data-navigo>Assignments</a>
  `;

  return wrap;
}
