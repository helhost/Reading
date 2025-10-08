import ChapterItem from "./ChapterItem.js";

export default function ArticleItem({
  article,              // { id, title, author?, location?, completed?, deadline? }
  onMeatballClick,      // ({ event, anchorEl, element, article }) => void
  onActionClick,        // (pillCtx, { article, element }) => void (same shape as ChapterItem ctx but +articleId)
} = {}) {
  if (!article || !Number.isInteger(article.id)) {
    throw new Error("ArticleItem: article with numeric id is required");
  }

  const li = document.createElement("li");
  li.className = "article-item book-item"; // reuse book-item spacing/positioning styles
  li.dataset.articleId = String(article.id);

  const title = document.createElement("div");
  title.className = "book-title"; // reuse typography
  title.textContent = article.title || "Untitled article";

  const meta = document.createElement("div");
  meta.className = "book-meta";
  const parts = [];
  if (article.author) parts.push(article.author);
  if (article.location) parts.push(article.location);
  meta.textContent = parts.length ? parts.join(" • ") : "Article";

  const actionWrap = document.createElement("div");
  actionWrap.className = "chapters"; // reuse pill layout gutter

  // Single pill (reuses ChapterItem for deadline/completed affordances)
  const pill = ChapterItem({
    index: 1,                       // arbitrary, we hide the label with CSS
    chapterId: null,                // not a chapter; we'll pass articleId in the ctx instead
    completed: !!article.completed,
    deadline: article.deadline ?? null,
    onClick: (ctx) => {
      // augment ctx with the real id
      onActionClick?.(
        { ...ctx, articleId: article.id },
        { article, element: li }
      );
    },
  });
  pill.classList.add("article-pill"); // font-size:0 to hide the "1"
  actionWrap.appendChild(pill);

  // Meatball menu (reuse book-menu styles)
  const more = document.createElement("button");
  more.type = "button";
  more.className = "book-menu-btn";
  more.textContent = "⋯";
  more.setAttribute("data-no-toggle", "");
  more.addEventListener("click", (e) => {
    e.stopPropagation();
    onMeatballClick?.({ event: e, anchorEl: more, element: li, article });
  });

  li.append(title, meta, actionWrap, more);

  // Tiny imperative API for callers
  li.setCompleted = (v) => pill.setCompleted?.(!!v);
  li.setDeadline = (tsOrNull) => pill.setDeadline?.(tsOrNull);
  li.getArticle = () => ({ ...article });

  return li;
}
