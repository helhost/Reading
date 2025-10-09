import CourseSection from "./CourseSection.js";
import articlesService from "../services/articles.js";
import courseService from "../services/courses.js";
import { OpenFormModal } from "../components/Form.js";
import { Toast } from "../components/Toast.js";
import ArticleItem from "../components/ArticleItem.js";
import openModal from "../components/Modal.js";
import openDatePicker from "../components/DatePicker.js";
import Button from "../components/Button.js";

// keep a handle to CourseSection API so we can move courses between my/remaining on unenroll
let courseSectionAPI = null;

export default function ArticlePage(myCourses = [], remainingCourses = [], universityId = null) {
  const section = CourseSection({
    myCourses,
    remainingCourses,
    universityId,
    renderBody: (course) => renderCourseArticles(course),
  });
  courseSectionAPI = section;
  return section;
}

function renderCourseArticles(course) {
  const wrap = document.createElement("div");
  wrap.className = "books"; // reuse spacing

  const list = document.createElement("ul");
  list.className = "books-list"; // reuse list layout
  wrap.appendChild(list);

  const loading = document.createElement("div");
  loading.className = "book-meta";
  loading.textContent = "Loading articles…";
  wrap.appendChild(loading);

  // Actions row: Add article + Leave course
  const actions = document.createElement("div");
  actions.className = "books-actions"; // reuse flex row styling

  const addBtn = Button({
    label: "＋ Add article",
    type: "primary",
    onClick: (e) => {
      e?.stopPropagation?.();
      openCreateArticleModal(course, {
        onCreated: (article) => {
          const item = ArticleItem({
            article,
            onMeatballClick: (ctx) => openArticleMenu(ctx),
            onActionClick: (ctx) => openArticleActions(ctx),
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

        if (courseSectionAPI?.getState && courseSectionAPI?.setLists) {
          const { my, remaining } = courseSectionAPI.getState();
          const myNext = my.filter((c) => c.id !== course.id);
          const remainingNext = [...remaining, course];
          courseSectionAPI.setLists(myNext, remainingNext);
        } else {
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

  // Load articles
  (async () => {
    try {
      const arts = await articlesService.listByCourse(course.id);
      list.innerHTML = "";

      if (!Array.isArray(arts) || arts.length === 0) {
        const empty = document.createElement("div");
        empty.className = "book-meta";
        empty.textContent = "No articles yet.";
        list.appendChild(empty);
        return;
      }

      for (const a of arts) {
        const item = ArticleItem({
          article: {
            id: Number(a.id),
            title: a.title,
            author: a.author ?? null,
            location: a.location ?? null,
            completed: !!a.completed,
            deadline: a.deadline ?? null,
          },
          onMeatballClick: (ctx) => openArticleMenu(ctx),
          onActionClick: (ctx) => openArticleActions(ctx),
        });
        list.appendChild(item);
      }
    } catch (err) {
      console.error("[articles] load failed:", err);
      Toast("error", err?.message || "Failed to load articles");
    } finally {
      loading.remove();
    }
  })();

  return wrap;
}

/* -------------------- meatball popover (Delete) -------------------- */

function openArticleMenu({ anchorEl, element: articleEl, article }) {
  const modal = openModal({
    anchorEl,
    placement: "bottom-end",
    offset: 8,
    cardClass: "bookmenu-popover modal-card--popover", // reuse compact menu style
  });

  const list = document.createElement("div");
  list.className = "bookmenu-list";

  const del = document.createElement("button");
  del.type = "button";
  del.className = "bookmenu-item bookmenu-item--danger";
  del.textContent = "Delete";
  del.addEventListener("click", async () => {
    try {
      await articlesService.delete(article.id);
      const ul = articleEl?.parentElement;
      articleEl?.remove();
      if (ul && ul.children.length === 0) {
        const empty = document.createElement("div");
        empty.className = "book-meta";
        empty.textContent = "No articles yet.";
        ul.appendChild(empty);
      }
      Toast("success", "Article deleted");
      modal.close();
    } catch (e) {
      const msg = e?.message || "";
      if (msg.includes("409") || /completed/i.test(msg)) {
        Toast("error", "Cannot delete: at least one person has completed it");
      } else {
        Toast("error", "Failed to delete article");
      }
      modal.close();
    }
  });

  list.appendChild(del);
  modal.setBody(list);
  return modal;
}

/* -------------------- action pill popover (complete / deadline) -------------------- */

function openArticleActions(ctx) {
  const { element: pillEl, completed, deadline, articleId } = ctx;

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
      // optimistic
      pillEl.setCompleted?.(!completed);
      if (!Number.isInteger(articleId) || articleId <= 0) {
        modal.close();
        return;
      }
      try {
        await articlesService.setProgress(articleId, !completed);
        modal.close();
      } catch (err) {
        pillEl.setCompleted?.(completed); // rollback
        console.error("Toggle article failed:", err);
        Toast("error", "Failed to update article");
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
      centered: true,
      initial: deadline ?? null,
      onPick: async (tsOrNull) => {
        // optimistic
        pillEl.setDeadline?.(tsOrNull);
        dead.textContent = tsOrNull ? formatDeadline(tsOrNull) : "No deadline";
        if (!Number.isInteger(articleId) || articleId <= 0) return;
        try {
          await articlesService.setDeadline(articleId, tsOrNull);
        } catch (err) {
          console.error("Set article deadline failed:", err);
          Toast("error", "Failed to set deadline");
        }
      },
    });
    setTimeout(() => modal.close(), 0);
  });

  row2.append(calBtn, dead);

  col.append(row1, row2);
  modal.setBody(col);

  return modal;
}

/* -------------------- create article -------------------- */

function openCreateArticleModal(course, { onCreated }) {
  OpenFormModal({
    title: "Create an article",
    submitLabel: "Create",
    fields: [
      {
        label: "Title",
        type: "string",
        required: true,
        name: "title",
        placeholder: "e.g., MapReduce: Simplified Data Processing on Large Clusters",
      },
      {
        label: "Author",
        type: "string",
        required: true,
        name: "author",
        placeholder: "e.g., Dean & Ghemawat",
      },
      {
        label: "Location (optional)",
        type: "string",
        required: false,
        name: "location",
        placeholder: "e.g., Library shelf B2 or URL",
      },
    ],
    onSubmit: async (data, { close }) => {
      try {
        const created = await articlesService.create({
          courseId: course.id,
          title: data.title,
          author: data.author,
          location: data.location || undefined,
        });

        Toast("success", `Added "${data.title}"`);
        onCreated?.({
          id: created.id,
          courseId: course.id,
          title: data.title,
          author: data.author,
          location: data.location || null,
          completed: false,
          deadline: null,
        });
        close();
      } catch (e) {
        Toast("error", e?.message || "Failed to create article");
      }
    },
  });
}

/* -------------------- utils -------------------- */

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
