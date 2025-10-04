import courseService from "./services/courses.js";

const courses = await courseService.getAll();

const container = document.createElement('div');
container.classList.add('course-container');
document.body.appendChild(container);

for (let course of courses) {
  const box = document.createElement('div');
  box.classList.add('course-box');

  const header = document.createElement('h2');
  header.classList.add('course-title');
  header.textContent = `${course.code}: ${course.name}`;

  const info = document.createElement('p');
  info.classList.add('course-info');
  info.textContent = `Term ${course.term}, ${course.year}`;

  box.appendChild(header);
  box.appendChild(info);

  if (Array.isArray(course.books) && course.books.length) {
    const booksWrap = document.createElement('div');
    booksWrap.classList.add('books');

    const list = document.createElement('ul');
    list.classList.add('books-list');

    course.books.forEach(b => {
      const li = document.createElement('li');
      li.classList.add('book-item');

      const title = document.createElement('div');
      title.classList.add('book-title');
      title.textContent = b.title;

      const meta = document.createElement('div');
      meta.classList.add('book-meta');
      meta.textContent = `${b.author} • ${b.numChapters} chapters`;

      li.appendChild(title);
      li.appendChild(meta);
      list.appendChild(li);
    });

    booksWrap.appendChild(list);
    box.appendChild(booksWrap);
  }

  // toggle visibility on click
  box.addEventListener('click', () => {
    box.classList.toggle('collapsed');
  });

  container.appendChild(box);
}
