import CourseSection from "./CourseSection.js";
import booksService from "../services/books.js";
import chaptersService from "../services/chapters.js";
import { OpenFormModal } from "../components/Form.js";
import { Toast } from "../components/Toast.js";
import BookItem from "../components/BookItem.js";
import openModal from "../components/Modal.js";
import openDatePicker from "../components/DatePicker.js";
import Button from "../components/Button.js";

export default function BookPage(myCourses = [], remainingCourses = []) {
  return CourseSection({
    myCourses,
    remainingCourses,
    renderBody: (course) => renderCourseBooks(course),
  });
}

function renderCourseBooks(course) {
  const wrap = document.createElement("div");
  wrap.className = "books";

  const list = document.createElement("ul");
  list.className = "books-list";
  wrap.appendChild(list);

  const loading = document.createElement("div");
  loading.className = "book-meta";
  loading.textContent = "Loading books…";
  wrap.appendChild(loading);

  // Add book
  const addBtn = document.createElement("button");
  addBtn.type = "button";
  addBtn.className = "add-book-btn";
  addBtn.textContent = "＋ Add book";
  addBtn.setAttribute("data-no-toggle", "");
  addBtn.addEventListener("click", (e) => {
    e.stopPropagation();
    openCreateBookModal(course, {
      onCreated: (book) => {
        const item = BookItem({
          book,
          chapters: synthesizeChapters(book.numChapters),
          onMeatballClick: defaultMeatballHandler,
          onChapterClick: (ctx) => openChapterActions(ctx),
        });
        list.appendChild(item);
      },
    });
  });
  wrap.appendChild(addBtn);

  // Load books (embedded chapters)
  (async () => {
    try {
      const books = await booksService.getByCourse(course.id);
      list.innerHTML = "";

      if (!Array.isArray(books) || books.length === 0) {
        const empty = document.createElement("div");
        empty.className = "book-meta";
        empty.textContent = "No books yet.";
        list.appendChild(empty);
        return;
      }

      for (const book of books) {
        const chapters = normalizeEmbeddedChapters(book);
        const item = BookItem({
          book,
          chapters,
          onMeatballClick: defaultMeatballHandler,
          onChapterClick: (ctx) => openChapterActions(ctx),
        });
        list.appendChild(item);
      }
    } catch (err) {
      console.error("[books] load failed:", err);
      Toast("error", err?.message || "Failed to load books");
    } finally {
      loading.remove();
    }
  })();

  return wrap;
}

/* -------------------- chapter actions -------------------- */

function openChapterActions(ctx) {
  const { element: pillEl, completed, deadline, chapterId } = ctx;

  const modal = openModal({
    anchorEl: pillEl,
    placement: "bottom-start",
    offset: 8,
    cardClass: "chapter-actions", // compact popover styling
  });

  // Column container (rows of actions)
  const col = document.createElement("div");
  col.className = "chapter-actions__col";

  // Row 1 — Complete toggle
  const row1 = document.createElement("div");
  row1.className = "chapter-actions__row";
  const completeBtn = Button({
    label: completed ? "Mark incomplete" : "Mark complete",
    type: completed ? "warn" : "success",
    onClick: async () => {
      // optimistic UI
      pillEl.setCompleted?.(!completed);

      // no id? UI-only until next refresh
      if (!Number.isInteger(chapterId) || chapterId <= 0) {
        modal.close();
        return;
      }
      try {
        await chaptersService.setProgress(chapterId, !completed);
        modal.close();
      } catch (err) {
        pillEl.setCompleted?.(completed); // rollback
        console.error("Toggle chapter failed:", err);
        Toast("error", "Failed to update chapter");
        modal.close();
      }
    },
  });
  row1.appendChild(completeBtn);

  // Row 2 — Calendar + deadline text
  const row2 = document.createElement("div");
  row2.className = "chapter-actions__row";

  const calBtn = document.createElement("button");
  calBtn.type = "button";
  calBtn.className = "chapter-actions__calendar";
  calBtn.textContent = "📅";
  calBtn.title = deadline ? `Deadline: ${formatDeadline(deadline)}` : "Set deadline";

  const dead = document.createElement("span");
  dead.className = "chapter-actions__deadline";
  dead.textContent = deadline ? formatDeadline(deadline) : "No deadline";

  calBtn.addEventListener("click", () => {
    // open date picker anchored to the calendar button
    openDatePicker({
      anchorEl: calBtn,
      initial: deadline ?? null,
      onPick: async (tsOrNull) => {
        // Optimistic UI update
        pillEl.setDeadline?.(tsOrNull);
        dead.textContent = tsOrNull ? formatDeadline(tsOrNull) : "No deadline";

        if (!Number.isInteger(chapterId) || chapterId <= 0) return;

        try {
          await chaptersService.setDeadline(chapterId, tsOrNull);
        } catch (err) {
          // rollback display if server fails (best-effort: we don't have previous text, just notify)
          console.error("Set deadline failed:", err);
          Toast("error", "Failed to set deadline");
        }
      },
      onClose: () => {
        // keep parent popover open; you can close it here if desired
      },
    });
  });

  row2.append(calBtn, dead);

  col.append(row1, row2);
  modal.setBody(col);

  return modal;
}

/* -------------------- utils -------------------- */

function defaultMeatballHandler({ anchorEl, book }) {
  console.log("book menu @", anchorEl, "for book", book);
}

function normalizeEmbeddedChapters(book) {
  if (Array.isArray(book.chapters) && book.chapters.length) {
    return book.chapters
      .map((r) => ({
        id: Number(r.id),
        index: Number(r.chapter_num),
        completed: !!r.completed,
        deadline: r.deadline ?? null,
      }))
      .sort((a, b) => a.index - b.index);
  }
  return synthesizeChapters(book.numChapters);
}

function synthesizeChapters(numChapters) {
  const n = Number(numChapters) || 0;
  return Array.from({ length: n }, (_, i) => ({
    id: null,
    index: i + 1,
    completed: false,
    deadline: null,
  }));
}

function formatDeadline(unixSeconds) {
  try {
    return new Date(unixSeconds * 1000).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "2-digit",
    });
  } catch {
    return String(unixSeconds);
  }
}

/* -------------------- create book -------------------- */

function openCreateBookModal(course, { onCreated }) {
  OpenFormModal({
    title: "Create a book",
    submitLabel: "Create",
    fields: [
      {
        label: "Title",
        type: "string",
        required: true,
        name: "title",
        placeholder: "e.g., Operating Systems: Three Easy Pieces",
      },
      {
        label: "Author",
        type: "string",
        required: true,
        name: "author",
        placeholder: "e.g., Remzi Arpaci-Dusseau",
      },
      {
        label: "Chapters",
        type: "int",
        required: true,
        name: "numChapters",
        placeholder: "e.g., 15",
      },
      {
        label: "Location (optional)",
        type: "string",
        required: false,
        name: "location",
        placeholder: "e.g., Library shelf A4",
      },
    ],
    onSubmit: async (data, { close }) => {
      if (data.numChapters <= 0) {
        Toast("warn", "Chapters must be a positive integer");
        return;
      }
      try {
        const created = await booksService.create({
          courseId: course.id,
          title: data.title,
          author: data.author,
          numChapters: data.numChapters,
          location: data.location || undefined,
        });

        Toast("success", `Added "${data.title}"`);
        onCreated?.({
          id: created.id,
          courseId: course.id,
          title: data.title,
          author: data.author,
          numChapters: data.numChapters,
          location: data.location || null,
        });
        close();
      } catch (e) {
        Toast("error", e?.message || "Failed to create book");
      }
    },
  });
}
