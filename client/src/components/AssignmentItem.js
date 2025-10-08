import ChapterItem from "./ChapterItem.js";

export default function AssignmentItem({
  assignment,           // { id, title, description?, completed?, deadline? }
  onMeatballClick,      // ({ event, anchorEl, element, assignment }) => void
  onActionClick,        // (pillCtx, { assignment, element }) => void
} = {}) {
  if (!assignment || !Number.isInteger(assignment.id)) {
    throw new Error("AssignmentItem: assignment with numeric id is required");
  }

  const li = document.createElement("li");
  li.className = "assignment-item book-item"; // reuse book-item spacing/positioning
  li.dataset.assignmentId = String(assignment.id);

  const title = document.createElement("div");
  title.className = "book-title"; // reuse typography
  title.textContent = assignment.title || "Untitled assignment";

  const meta = document.createElement("div");
  meta.className = "book-meta";
  meta.textContent = assignment.description?.trim()
    ? assignment.description.trim()
    : "Assignment";

  const actionWrap = document.createElement("div");
  actionWrap.className = "chapters"; // reuse pill layout gutter

  // Single pill (hide the numeric label via CSS)
  const pill = ChapterItem({
    index: 1,
    chapterId: null,
    completed: !!assignment.completed,
    deadline: assignment.deadline ?? null,
    onClick: (ctx) => {
      onActionClick?.(
        { ...ctx, assignmentId: assignment.id },
        { assignment, element: li }
      );
    },
  });
  pill.classList.add("assignment-pill");
  actionWrap.appendChild(pill);

  // Meatball menu
  const more = document.createElement("button");
  more.type = "button";
  more.className = "book-menu-btn";
  more.textContent = "⋯";
  more.setAttribute("data-no-toggle", "");
  more.addEventListener("click", (e) => {
    e.stopPropagation();
    onMeatballClick?.({ event: e, anchorEl: more, element: li, assignment });
  });

  li.append(title, meta, actionWrap, more);

  // Tiny imperative API
  li.setCompleted = (v) => pill.setCompleted?.(!!v);
  li.setDeadline = (tsOrNull) => pill.setDeadline?.(tsOrNull);
  li.getAssignment = () => ({ ...assignment });

  return li;
}
