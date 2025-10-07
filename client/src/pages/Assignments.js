import CourseSection from "./CourseSection.js";

export default function Assignments(myCourses = [], remainingCourses = []) {
  return CourseSection({
    myCourses,
    remainingCourses,
    renderBody: (course) => {
      const wrap = document.createElement("div");
      wrap.className = "assignments";

      const ul = document.createElement("ul");
      ul.className = "assignments-list";
      wrap.appendChild(ul);

      const addBtn = document.createElement("button");
      addBtn.type = "button";
      addBtn.className = "add-assignment-btn";
      addBtn.textContent = "＋ Add assignment";
      addBtn.setAttribute("data-no-toggle", "");
      addBtn.addEventListener("click", (e) => e.stopPropagation());
      wrap.appendChild(addBtn);

      return wrap;
    },
  });
}
