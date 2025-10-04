import courseService from "./services/courses.js";
import bookService from "./services/books.js";
import { openBookForm } from "./bookForm.js";

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
  box.setAttribute('aria-expanded', 'true'); // default open

  // header
  const header = document.createElement('h2');
  header.classList.add('course-title');
  header.textContent = `${course.code}: ${course.name}`;

  const info = document.createElement('p');
  info.classList.add('course-info');
  info.textContent = `Term ${course.term}, ${course.year}`;

  const indicator = drawIndicator();

  box.append(header, info, indicator);

  // books (optional)
  if (Array.isArray(course.books) && course.books.length) {
    const booksWrap = document.createElement('div');
    booksWrap.classList.add('books');

    const list = document.createElement('ul');
    list.classList.add('books-list');

    for (const book of course.books) {
      list.appendChild(drawBook(book));
    }

    booksWrap.appendChild(list);
    box.appendChild(booksWrap);
  }

  // + Add book
  box.appendChild(drawAddBookButton(course.id));

  // toggle open/closed
  box.addEventListener('click', () => {
    const isCollapsed = box.classList.toggle('collapsed');
    box.setAttribute('aria-expanded', String(!isCollapsed));
  });

  return box;
}

function drawAddBookButton(courseId) {
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'add-book-btn books-action';
  btn.textContent = '＋ Add book';
  btn.addEventListener('click', (e) => {
    e.stopPropagation();
    handleAddBook(courseId);
  });
  return btn;
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

function handleAddBook(courseId) {
  openBookForm(courseId, {
    onSubmit: async (data, { close }) => {
      // Coerce & validate
      const numChapters = Number(data.numChapters);
      if (!Number.isFinite(numChapters) || numChapters < 0) {
        console.error('Invalid numChapters:', data.numChapters);
        return; // or show a message
      }

      const payload = {
        courseId,
        title: data.title,
        author: data.author,
        numChapters,                     // <- number, not string
        ...(data.link && data.link.trim() ? { link: data.link.trim() } : {}),
      };

      try {
        const created = await bookService.create(payload);
        console.log('created book:', created);
        close();
      } catch (err) {
        console.error('Create failed:', err);
      }
    }
  });
}
