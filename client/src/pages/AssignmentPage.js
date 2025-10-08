import CourseSection from "./CourseSection.js";
import assignmentsService from "../services/assignments.js";
import courseService from "../services/courses.js";
import { OpenFormModal } from "../components/Form.js";
import { Toast } from "../components/Toast.js";
import AssignmentItem from "../components/AssignmentItem.js";
import openModal from "../components/Modal.js";
import openDatePicker from "../components/DatePicker.js";
import Button from "../components/Button.js";

// keep a handle to CourseSection API so we can move courses between my/remaining on unenroll
let courseSectionAPI = null;

export default function AssignmentPage(myCourses = [], remainingCourses = []) {
  const section = CourseSection({
    myCourses,
    remainingCourses,
    renderBody: (course) => renderCourseAssignments(course),
  });
  courseSectionAPI = section;
  return section;
}

function renderCourseAssignments(course) {
  const wrap = document.createElement("div");
  wrap.className = "books"; // reuse spacing

  const list = document.createElement("ul");
  list.className = "books-list"; // reuse list layout
  wrap.appendChild(list);

  const loading = document.createElement("div");
  loading.className = "book-meta";
  loading.textContent = "Loading assignments…";
  wrap.appendChild(loading);

  // Actions row: Add assignment + Leave course
  const actions = document.createElement("div");
  actions.className = "books-actions";

  const addBtn = Button({
    label: "＋ Add assignment",
    type: "primary",
    onClick: (e) => {
      e?.stopPropagation?.();
      openCreateAssignmentModal(course, {
        onCreated: (assignment) => {
          const item = AssignmentItem({
            assignment,
            onMeatballClick: (ctx) => openAssignmentMenu(ctx),
            onActionClick: (ctx) => openAssignmentActions(ctx),
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

  // Load assignments
  (async () => {
    try {
      const rows = await assignmentsService.listByCourse(course.id);
      list.innerHTML = "";

      if (!Array.isArray(rows) || rows.length === 0) {
        const empty = document.createElement("div");
        empty.className = "book-meta";
        empty.textContent = "No assignments yet.";
        list.appendChild(empty);
        return;
      }

      for (const a of rows) {
        const item = AssignmentItem({
          assignment: {
            id: Number(a.id),
            title: a.title,
            description: a.description ?? null,
            completed: !!a.completed,
            deadline: a.deadline ?? null,
          },
          onMeatballClick: (ctx) => openAssignmentMenu(ctx),
          onActionClick: (ctx) => openAssignmentActions(ctx),
        });
        list.appendChild(item);
      }
    } catch (err) {
      console.error("[assignments] load failed:", err);
      Toast("error", err?.message || "Failed to load assignments");
    } finally {
      loading.remove();
    }
  })();

  return wrap;
}

/* -------------------- meatball popover (Delete) -------------------- */

function openAssignmentMenu({ anchorEl, element: el, assignment }) {
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
      await assignmentsService.delete(assignment.id);
      const ul = el?.parentElement;
      el?.remove();
      if (ul && ul.children.length === 0) {
        const empty = document.createElement("div");
        empty.className = "book-meta";
        empty.textContent = "No assignments yet.";
        ul.appendChild(empty);
      }
      Toast("success", "Assignment deleted");
      modal.close();
    } catch (e) {
      const msg = e?.message || "";
      if (msg.includes("409") || /completed/i.test(msg)) {
        Toast("error", "Cannot delete: at least one person has completed it");
      } else {
        Toast("error", "Failed to delete assignment");
      }
      modal.close();
    }
  });

  list.appendChild(del);
  modal.setBody(list);
  return modal;
}

/* -------------------- action pill popover (complete / deadline) -------------------- */

function openAssignmentActions(ctx) {
  const { element: pillEl, completed, deadline, assignmentId } = ctx;

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
      if (!Number.isInteger(assignmentId) || assignmentId <= 0) {
        modal.close();
        return;
      }
      try {
        await assignmentsService.setProgress(assignmentId, !completed);
        modal.close();
      } catch (err) {
        pillEl.setCompleted?.(completed); // rollback
        console.error("Toggle assignment failed:", err);
        Toast("error", "Failed to update assignment");
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
        // optimistic
        pillEl.setDeadline?.(tsOrNull);
        dead.textContent = tsOrNull ? formatDeadline(tsOrNull) : "No deadline";
        if (!Number.isInteger(assignmentId) || assignmentId <= 0) return;
        try {
          await assignmentsService.setDeadline(assignmentId, tsOrNull);
        } catch (err) {
          console.error("Set assignment deadline failed:", err);
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

/* -------------------- create assignment -------------------- */

function openCreateAssignmentModal(course, { onCreated }) {
  OpenFormModal({
    title: "Create an assignment",
    submitLabel: "Create",
    fields: [
      {
        label: "Title",
        type: "string",
        required: true,
        name: "title",
        placeholder: "e.g., Project 1: Filesystem",
      },
      {
        label: "Description (optional)",
        type: "string",
        required: false,
        name: "description",
        placeholder: "e.g., due next Friday; implement journaling",
      },
    ],
    onSubmit: async (data, { close }) => {
      try {
        const created = await assignmentsService.create({
          courseId: course.id,
          title: data.title,
          description: data.description || undefined,
        });

        Toast("success", `Added "${data.title}"`);
        onCreated?.({
          id: created.id,
          courseId: course.id,
          title: data.title,
          description: data.description || null,
          completed: false,
          deadline: null,
        });
        close();
      } catch (e) {
        Toast("error", e?.message || "Failed to create assignment");
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
