import CourseSection from "./CourseSection.js";

export default function Books(myCourses = [], remainingCourses = []) {
  return CourseSection({
    myCourses,
    remainingCourses,
    renderBody: (course) => {
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
}
