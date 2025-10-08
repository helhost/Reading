import ExpandableCard from "../components/ExpandableCard.js";
import Button from "../components/Button.js";
import { OpenSearchListModal } from "../components/SearchListModal.js";
import { OpenFormModal } from "../components/Form.js";
import { Toast } from "../components/Toast.js";
import courseService from "../services/courses.js";

export default function CourseSection({
  myCourses = [],
  remainingCourses = [],
  renderBody,
  slotSelector = ".page-body",
  addCourseButtonLabel = "＋ Add course",
  universityId = null,
} = {}) {
  if (typeof renderBody !== "function") {
    throw new Error("CourseSection: renderBody(course) is required");
  }

  const slot = document.querySelector(slotSelector);
  if (!slot) return;

  // local, mutable state
  let my = Array.isArray(myCourses) ? [...myCourses] : [];
  let remaining = Array.isArray(remainingCourses) ? [...remainingCourses] : [];

  const uniId = universityId ?? null;

  // reset UI
  slot.innerHTML = "";

  // container
  const list = document.createElement("div");
  list.className = "course-container";

  // empty state (if no enrolled)
  let emptyCard = null;
  if (my.length === 0) {
    emptyCard = document.createElement("div");
    emptyCard.className = "card";
    emptyCard.textContent = "You are not enrolled in any courses.";
    list.appendChild(emptyCard);
  } else {
    my.forEach((c) => list.appendChild(makeCourseCard(c)));
  }

  // footer with enroll/create
  const footer = document.createElement("div");
  footer.className = "uni-footer";

  const addBtn = Button({
    label: addCourseButtonLabel,
    type: "primary",
    onClick: openEnrollModal,
  });
  footer.appendChild(addBtn);

  slot.append(list, footer);
  window.router?.updatePageLinks?.();

  // ---------- helpers ----------

  function makeCourseCard(c) {
    const card = ExpandableCard({
      title: `${c.code}: ${c.name}`,
      subtitle: `Term ${c.term}, ${c.year}`,
      initiallyOpen: false,
      content: () => {
        const node = renderBody(c);
        return node instanceof Node ? node : document.createTextNode("");
      },
    });
    card.dataset.courseId = c.id;
    return card;
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
        } catch (e) {
          Toast("error", e?.message || "Failed to enroll");
        }
      },
      footerLabel: "Create a new course",
      onFooterClick: openCreateCourseModal,
    });
  }

  function openCreateCourseModal() {
    if (!uniId) {
      Toast("error", "Cannot create: missing university context");
      return;
    }

    OpenFormModal({
      title: "Create a course",
      submitLabel: "Create",
      fields: [
        { label: "Code", type: "string", required: true, name: "code", placeholder: "e.g., CS101" },
        { label: "Name", type: "string", required: true, name: "name", placeholder: "e.g., Intro to CS" },
        { label: "Year", type: "int", required: true, name: "year", initial: new Date().getFullYear() },
        { label: "Term", type: "int", required: true, name: "term", placeholder: "e.g., 1" },
      ],
      onSubmit: async (data, { close }) => {
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
            universityId: uniId,
            code: data.code,
            name: data.name,
            year: data.year,
            term: data.term,
          });

          await courseService.enroll(created.id);
          remaining = remaining.filter((x) => x.id !== created.id);

          addCourseToUI({
            id: created.id,
            universityId: uniId,
            code: created.code,
            name: created.name,
            year: created.year,
            term: created.term,
          });

          Toast("success", `Created & enrolled in ${created.code}: ${created.name}`);
          close();
        } catch (e) {
          Toast("error", e?.message || "Failed to create course");
        }
      },
    });
  }

  // expose a tiny API if the caller wants to mutate later
  return {
    getState: () => ({ my: [...my], remaining: [...remaining], universityId: uniId }),
    addCourseToUI,
    setLists(nextMy = [], nextRemaining = []) {
      my = Array.isArray(nextMy) ? [...nextMy] : [];
      remaining = Array.isArray(nextRemaining) ? [...nextRemaining] : [];
      // re-render minimal (clear list only)
      list.innerHTML = "";
      if (my.length === 0) {
        emptyCard = document.createElement("div");
        emptyCard.className = "card";
        emptyCard.textContent = "You are not enrolled in any courses.";
        list.appendChild(emptyCard);
      } else {
        emptyCard = null;
        my.forEach((c) => list.appendChild(makeCourseCard(c)));
      }
    },
  };
}
