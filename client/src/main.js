import courseService from "./services/courses.js";
import bookService from "./services/books.js";

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

  // state for this book (used by click handler)
  li.dataset.bookId = book.id;
  li.dataset.completed = String(book.completedChapters || 0);
  li.dataset.numChapters = String(book.numChapters);

  const title = document.createElement('div');
  title.classList.add('book-title');
  title.textContent = book.title;

  const meta = document.createElement('div');
  meta.classList.add('book-meta');
  meta.textContent = `${book.author} • ${book.numChapters} chapters`;

  const chapters = drawChapters(book); // unchanged call-site
  li.append(title, meta, chapters);
  return li;
}


function drawChapters(book) {
  const wrap = document.createElement('div');
  wrap.classList.add('chapters');

  const current = book.completedChapters || 0;

  for (let i = 1; i <= book.numChapters; i++) {
    const ch = document.createElement('div');
    ch.classList.add('chapter-box');
    ch.textContent = i;
    if (i <= current) ch.classList.add('completed');

    ch.addEventListener('click', (e) => {
      e.stopPropagation();
      const bookEl = e.currentTarget.closest('.book-item');
      handleChapterClick(bookEl, book.id, i); // pass DOM + ids
    });

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


function applyCompletedUI(bookEl, completed) {
  const boxes = bookEl.querySelectorAll('.chapter-box');
  boxes.forEach((box, idx) => {
    if (idx < completed) box.classList.add('completed');
    else box.classList.remove('completed');
  });
  bookEl.dataset.completed = String(completed);
}

async function handleChapterClick(bookEl, bookId, n) {
  const current = Number(bookEl.dataset.completed || 0);
  const next = (n === current) ? 0 : n;

  // optimistic UI
  applyCompletedUI(bookEl, next);

  try {
    const updated = await bookService.updateCompletedChapters(bookId, next);
    const confirmed = Number(
      (updated && updated.completedChapters) != null
        ? updated.completedChapters
        : next
    );
    applyCompletedUI(bookEl, confirmed); // ensure UI matches server
  } catch (err) {
    // rollback
    applyCompletedUI(bookEl, current);
    console.error('Failed to update completedChapters', err);
  }
}
