import ExpandableCard from "../components/ExpandableCard.js";
import Button from "../components/Button.js";
import { OpenSearchListModal } from "../components/SearchListModal.js";
import { OpenFormModal } from "../components/Form.js";
import { Toast } from "../components/Toast.js";
import courseService from "../services/courses.js";

export default function Books(myCourses = [], remainingCourses = []) {
  const slot = document.querySelector(".page-body");
  if (!slot) return;

  // local, mutable copies
  let my = Array.isArray(myCourses) ? [...myCourses] : [];
  let remaining = Array.isArray(remainingCourses) ? [...remainingCourses] : [];

  const universityId =
    my?.[0]?.universityId ?? remaining?.[0]?.universityId ?? null;

  // reset
  slot.innerHTML = "";

  // containers
  const list = document.createElement("div");
  list.className = "course-container";

  // optional empty state
  let emptyCard = null;
  if (my.length === 0) {
    emptyCard = document.createElement("div");
    emptyCard.className = "card";
    emptyCard.textContent = "You are not enrolled in any courses.";
    list.appendChild(emptyCard);
  } else {
    my.forEach((c) => list.appendChild(makeCourseCard(c)));
  }

  const footer = document.createElement("div");
  footer.className = "uni-footer";

  const addCourseBtn = Button({
    label: "＋ Add course",
    type: "primary",
    onClick: openEnrollModal,
    disabled: remaining.length === 0 && !universityId, // allow create if we know the uni
  });
  footer.appendChild(addCourseBtn);

  slot.append(list, footer);
  window.router?.updatePageLinks?.();

  // ---------- helpers ----------

  function makeCourseCard(c) {
    const card = ExpandableCard({
      title: `${c.code}: ${c.name}`,
      subtitle: `Term ${c.term}, ${c.year}`,
      initiallyOpen: false,
      content: () => {
        const wrap = document.createElement("div");
        wrap.className = "books";

        const ul = document.createElement("ul");
        ul.className = "books-list";
        wrap.appendChild(ul);

        const addBtn = document.createElement("button");
        addBtn.type = "button";
        addBtn.className = "add-book-btn";
        addBtn.textContent = "＋ Add book";
        addBtn.setAttribute("data-no-toggle", "");
        addBtn.addEventListener("click", (e) => e.stopPropagation());
        wrap.appendChild(addBtn);

        return wrap;
      },
    });
    card.dataset.courseId = c.id;
    return card;
  }

  function reflectAddButton() {
    addCourseBtn.disabled = remaining.length === 0 && !universityId;
  }

  function addCourseToUI(c) {
    if (emptyCard) {
      emptyCard.remove();
      emptyCard = null;
    }
    my.push(c);
    list.appendChild(makeCourseCard(c));
    window.router?.updatePageLinks?.();
  }

  // ---------- actions ----------

  function openEnrollModal() {
    OpenSearchListModal({
      title: "Enroll in a course",
      items: remaining,
      getTitle: (c) => `${c.code}: ${c.name} — Term ${c.term}, ${c.year}`,
      actionLabel: "Enroll",
      onPick: async (c) => {
        try {
          await courseService.enroll(c.id);
          Toast("success", `Enrolled in ${c.code}: ${c.name}`);

          remaining = remaining.filter((x) => x.id !== c.id);
          addCourseToUI(c);
          reflectAddButton();
        } catch (e) {
          Toast("error", e?.message || "Failed to enroll");
        }
      },
      footerLabel: "Create a new course",
      onFooterClick: openCreateCourseModal, // <-- now uses Form.js
    });
  }

  function openCreateCourseModal() {
    if (!universityId) {
      Toast("error", "Cannot create: missing university context");
      return;
    }

    OpenFormModal({
      title: "Create a course",
      submitLabel: "Create",
      fields: [
        { label: "Code", type: "string", required: true, name: "code", placeholder: "e.g., CS101" },
        { label: "Name", type: "string", required: true, name: "name", placeholder: "e.g., Intro to CS" },
        { label: "Term", type: "int", required: true, name: "term", placeholder: "e.g., 1" },
        { label: "Year", type: "int", required: true, name: "year", initial: new Date().getFullYear() },
      ],
      onSubmit: async (data, { close }) => {
        // extra guards beyond required/int
        if (data.year < 1900) {
          Toast("warn", "Please enter a realistic year (>= 1900)");
          return;
        }
        if (data.term <= 0) {
          Toast("warn", "Term must be a positive number");
          return;
        }

        try {
          const created = await courseService.create({
            universityId,
            code: data.code,
            name: data.name,
            year: data.year,
            term: data.term,
          });

          await courseService.enroll(created.id);

          remaining = remaining.filter((x) => x.id !== created.id);

          addCourseToUI({
            id: created.id,
            universityId,
            code: created.code,
            name: created.name,
            year: created.year,
            term: created.term,
          });

          Toast("success", `Created & enrolled in ${created.code}: ${created.name}`);
          reflectAddButton();
          close();
        } catch (e) {
          Toast("error", e?.message || "Failed to create course");
          // keep modal open for correction
        }
      },
    });
  }
}
