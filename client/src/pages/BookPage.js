import CourseSection from "./CourseSection.js";
import booksService from "../services/books.js";
import chaptersService from "../services/chapters.js";
import courseService from "../services/courses.js";
import { OpenFormModal } from "../components/Form.js";
import { Toast } from "../components/Toast.js";
import BookItem from "../components/BookItem.js";
import openModal from "../components/Modal.js";
import openDatePicker from "../components/DatePicker.js";
import Button from "../components/Button.js";

// Will hold the CourseSection API after mount
let courseSectionAPI = null;

export default function BookPage(myCourses = [], remainingCourses = []) {
  // renderBody closes over courseSectionAPI; it's assigned right after mount.
  const section = CourseSection({
    myCourses,
    remainingCourses,
    renderBody: (course) => renderCourseBooks(course),
  });
  courseSectionAPI = section; // now available to leave handler
  return section;
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

  // Actions row: Add book + Leave course (both via Button component)
  const actions = document.createElement("div");
  actions.className = "books-actions";

  const addBtn = Button({
    label: "＋ Add book",
    type: "primary",
    onClick: (e) => {
      e?.stopPropagation?.();
      openCreateBookModal(course, {
        onCreated: (book) => {
          const item = BookItem({
            book,
            chapters: synthesizeChapters(book.numChapters),
            onMeatballClick: (ctx) => openBookMenu(ctx),
            onChapterClick: (ctx) => openChapterActions(ctx),
          });
          list.appendChild(item);
        },
      });
    },
  });
  addBtn.setAttribute("data-no-toggle", "");

  const leaveBtn = Button({
    label: "Leave course",
    type: "danger",
    onClick: async (e) => {
      e?.stopPropagation?.();
      if (!confirm(`Leave ${course.code}: ${course.name}?`)) return;
      try {
        await courseService.unenroll(course.id);
        Toast("success", `Left ${course.code}`);

        // Preferred: update CourseSection state so remaining/enroll modal refreshes
        if (courseSectionAPI?.getState && courseSectionAPI?.setLists) {
          const { my, remaining } = courseSectionAPI.getState();
          const myNext = my.filter((c) => c.id !== course.id);
          const remainingNext = [...remaining, course];
          courseSectionAPI.setLists(myNext, remainingNext);
        } else {
          // Fallback: remove just this card (won’t update remaining list)
          const cardEl = wrap.closest(".exp-card");
          const container = cardEl?.parentElement;
          cardEl?.remove();
          if (container && !container.querySelector(".exp-card")) {
            const empty = document.createElement("div");
            empty.className = "card";
            empty.textContent = "You are not enrolled in any courses.";
            container.appendChild(empty);
          }
        }
      } catch (err) {
        console.error("Unenroll failed:", err);
        Toast("error", err?.message || "Failed to leave course");
      }
    },
  });
  leaveBtn.setAttribute("data-no-toggle", "");

  actions.append(addBtn, leaveBtn);
  wrap.appendChild(actions);

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
          onMeatballClick: (ctx) => openBookMenu(ctx),
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

/* -------------------- meatball popover -------------------- */

function openBookMenu({ anchorEl, element: bookItemEl, book }) {
  const modal = openModal({
    anchorEl,
    placement: "bottom-end",
    offset: 8,
    cardClass: "bookmenu-popover modal-card--popover",
  });

  const list = document.createElement("div");
  list.className = "bookmenu-list";

  const del = document.createElement("button");
  del.type = "button";
  del.className = "bookmenu-item bookmenu-item--danger";
  del.textContent = "Delete";
  del.addEventListener("click", async () => {
    try {
      await booksService.delete(book.id);
      const ul = bookItemEl?.parentElement;
      bookItemEl?.remove();
      if (ul && ul.children.length === 0) {
        const empty = document.createElement("div");
        empty.className = "book-meta";
        empty.textContent = "No books yet.";
        ul.appendChild(empty);
      }
      Toast("success", "Book deleted");
      modal.close();
    } catch (e) {
      const msg = e?.message || "";
      if (msg.includes("409") || /completed/i.test(msg)) {
        Toast("error", "Cannot delete: at least one person has completed a chapter in this book");
      } else {
        Toast("error", "Failed to delete book");
      }
      modal.close();
    }
  });

  list.appendChild(del);
  modal.setBody(list);
  return modal;
}

/* -------------------- chapter actions -------------------- */

function openChapterActions(ctx) {
  const { element: pillEl, completed, deadline, chapterId } = ctx;

  const modal = openModal({
    anchorEl: pillEl,
    placement: "bottom-start",
    offset: 8,
    cardClass: "chapter-actions modal-card--popover",
  });

  const col = document.createElement("div");
  col.className = "chapter-actions__col";

  // Row 1 — Complete toggle
  const row1 = document.createElement("div");
  row1.className = "chapter-actions__row";
  const completeBtn = Button({
    label: completed ? "Mark incomplete" : "Mark complete",
    type: completed ? "warn" : "success",
    onClick: async () => {
      pillEl.setCompleted?.(!completed); // optimistic
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
    openDatePicker({
      anchorEl: calBtn,
      initial: deadline ?? null,
      onPick: async (tsOrNull) => {
        pillEl.setDeadline?.(tsOrNull); // optimistic
        dead.textContent = tsOrNull ? formatDeadline(tsOrNull) : "No deadline";
        if (!Number.isInteger(chapterId) || chapterId <= 0) return;
        try {
          await chaptersService.setDeadline(chapterId, tsOrNull);
        } catch (err) {
          console.error("Set deadline failed:", err);
          Toast("error", "Failed to set deadline");
        }
      },
    });
  });

  row2.append(calBtn, dead);

  col.append(row1, row2);
  modal.setBody(col);

  return modal;
}

/* -------------------- utils -------------------- */

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
