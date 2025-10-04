import courseService from "./services/courses.js";

const courses = await courseService.getAll();


const container = document.createElement('div');
container.classList.add('course-container');
document.body.appendChild(container);

for (let course of courses) {
  const box = drawCourse(course);
  container.appendChild(box);
}


function drawCourse(course) {
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
    box.classList.add('has-books');
    const indicator = drawIndicator();
    const booksWrap = document.createElement('div');
    booksWrap.classList.add('books');

    const list = document.createElement('ul');
    list.classList.add('books-list');

    for (let book of course.books) {
      const li = drawBook(book);
      list.appendChild(li);
    }

    booksWrap.appendChild(list);
    box.appendChild(indicator);
    box.appendChild(booksWrap);
  }

  box.addEventListener('click', () => box.classList.toggle('collapsed'));
  return box;
}

function drawBook(book) {
  const li = document.createElement('li');
  li.classList.add('book-item');

  const title = document.createElement('div');
  title.classList.add('book-title');
  title.textContent = book.title;

  const meta = document.createElement('div');
  meta.classList.add('book-meta');
  meta.textContent = `${book.author} • ${book.numChapters} chapters`;

  const chapters = drawChapters(book);
  li.append(title, meta, chapters);
  return li;
}

function drawChapters(book) {
  const wrap = document.createElement('div');
  wrap.classList.add('chapters');
  for (let i = 1; i <= book.numChapters; i++) {
    const ch = document.createElement('div');
    ch.classList.add('chapter-box');
    ch.textContent = i;
    if (i <= (book.completedChapters || 0)) ch.classList.add('completed');
    wrap.appendChild(ch);
  }
  return wrap;
}

function drawIndicator() {
  const span = document.createElement('span');
  span.className = 'toggle-indicator';
  span.textContent = '▾';
  return span;
}
