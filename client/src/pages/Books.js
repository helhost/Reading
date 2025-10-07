import ExpandableCard from "../components/ExpandableCard.js";

export default function Books(myCourses = [], remainingCourses = []) {
  const content = document.querySelector(".page-body");
  if (!content) return;

  const my = Array.isArray(myCourses) ? myCourses : [];

  // reset the slot
  content.innerHTML = "";

  // empty state
  if (my.length === 0) {
    const empty = document.createElement("div");
    empty.className = "card";
    empty.textContent = "You are not enrolled in any courses.";
    content.appendChild(empty);
    return;
  }

  // list container (reuses your existing layout spacing)
  const list = document.createElement("div");
  list.className = "course-container";

  for (const c of my) {
    const title = `${c.code}: ${c.name}`;
    const subtitle = `Term ${c.term}, ${c.year}`;

    const card = ExpandableCard({
      title,
      subtitle,
      initiallyOpen: false,
      content: () => {
        // Placeholder book section; we’ll populate with books later.
        const wrap = document.createElement("div");
        wrap.className = "books";

        const ul = document.createElement("ul");
        ul.className = "books-list";
        wrap.appendChild(ul);

        // Non-functional for now; prevents collapsing when clicked.
        const addBtn = document.createElement("button");
        addBtn.type = "button";
        addBtn.className = "add-book-btn";
        addBtn.textContent = "＋ Add book";
        addBtn.setAttribute("data-no-toggle", "");
        addBtn.addEventListener("click", (e) => e.stopPropagation());
        wrap.appendChild(addBtn);

        return wrap;
      },
      onToggle: () => { },
    });

    // stash id if needed later
    card.dataset.courseId = c.id;

    list.appendChild(card);
  }

  content.appendChild(list);

  // If any internal links appear later, let Navigo bind them.
  window.router?.updatePageLinks?.();
}
