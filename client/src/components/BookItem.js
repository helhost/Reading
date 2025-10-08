import ChapterItem from "./ChapterItem.js";

export default function BookItem({
  book,                 // { id, title, author, numChapters }
  chapters = [],        // [{ id, index, completed, deadline }]
  onMeatballClick,      // ({ event, anchorEl, element, book }) => void
  onChapterClick,       // (chapterCtx, { book, element }) => void
} = {}) {
  if (!book || !Number.isInteger(book.id)) {
    throw new Error("BookItem: book with numeric id is required");
  }

  const li = document.createElement("li");
  li.className = "book-item";
  li.dataset.bookId = String(book.id);

  const title = document.createElement("div");
  title.className = "book-title";
  title.textContent = book.title || "Untitled book";

  const meta = document.createElement("div");
  meta.className = "book-meta";
  const n = Number(book.numChapters) || 0;
  meta.textContent = `${book.author ?? "Unknown author"} • ${n} chapter${n === 1 ? "" : "s"}`;

  const chaptersWrap = document.createElement("div");
  chaptersWrap.className = "chapters";
  chaptersWrap.setAttribute("data-no-toggle", "");

  const more = document.createElement("button");
  more.type = "button";
  more.className = "book-menu-btn";
  more.textContent = "⋯";
  more.setAttribute("data-no-toggle", "");
  more.addEventListener("click", (e) => {
    e.stopPropagation();
    onMeatballClick?.({ event: e, anchorEl: more, element: li, book });
  });

  li.append(title, meta, chaptersWrap, more);

  // render chapters
  const renderChapters = (list) => {
    chaptersWrap.innerHTML = "";
    for (const ch of list) {
      const node = ChapterItem({
        index: ch.index,
        chapterId: ch.id ?? null,
        completed: !!ch.completed,
        deadline: ch.deadline ?? null,
        onClick: (ctx) => onChapterClick?.(ctx, { book, element: li }),
      });
      chaptersWrap.appendChild(node);
    }
  };

  if (Array.isArray(chapters) && chapters.length) {
    renderChapters([...chapters].sort((a, b) => a.index - b.index));
  } else {
    const nn = Number(book.numChapters) || 0;
    const synthetic = Array.from({ length: nn }, (_, i) => ({
      id: null,
      index: i + 1,
      completed: false,
      deadline: null,
    }));
    renderChapters(synthetic);
  }

  // tiny API
  li.setChapters = (list) => {
    if (!Array.isArray(list)) return;
    renderChapters([...list].sort((a, b) => a.index - b.index));
  };
  li.updateChapter = (index, patch = {}) => {
    const pill = [...chaptersWrap.children].find((n) => Number(n.textContent) === Number(index));
    if (!pill) return;
    if ("completed" in patch && typeof pill.setCompleted === "function") {
      pill.setCompleted(!!patch.completed);
    }
    if ("deadline" in patch && typeof pill.setDeadline === "function") {
      pill.setDeadline(patch.deadline ?? null);
    }
  };
  li.getBook = () => ({ ...book });

  return li;
}
