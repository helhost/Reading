import CourseSection from "./CourseSection.js";

export default function ArticlePage(myCourses = [], remainingCourses = []) {
  return CourseSection({
    myCourses,
    remainingCourses,
    renderBody: (course) => {
      const wrap = document.createElement("div");
      wrap.className = "articles";

      const ul = document.createElement("ul");
      ul.className = "articles-list";
      wrap.appendChild(ul);

      const addBtn = document.createElement("button");
      addBtn.type = "button";
      addBtn.className = "add-article-btn";
      addBtn.textContent = "＋ Add article";
      addBtn.setAttribute("data-no-toggle", "");
      addBtn.addEventListener("click", (e) => e.stopPropagation());
      wrap.appendChild(addBtn);

      return wrap;
    },
  });
}
