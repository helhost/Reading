import courseService from "./services/courses.js";
import bookService from "./services/books.js";
import { openBookForm } from "./bookForm.js";
import { openBookMenu } from './bookMenu.js';
import { openCourseForm } from "./courseForm.js";

const courses = await courseService.getAll();

const container = document.createElement('div');
container.classList.add('course-container');
document.body.appendChild(container);

for (let course of courses) {
  const box = drawCourse(course);
  container.appendChild(box);
}
addCourseBtn();

// ---- helpers for completedChapters ([]int) ----

function coerceArray(v) {
  if (!v) return [];
  if (Array.isArray(v)) return v.map(Number).filter(n => Number.isFinite(n));
  if (typeof v === 'string') {
    if (v.trim() === '') return [];
    return v.split(',').map(s => Number(s.trim())).filter(n => Number.isFinite(n));
  }
  return [];
}

function getCompletedFromEl(bookEl) {
  try {
    const raw = bookEl.dataset.completedChapters || "[]";
    // dataset stores JSON string
    const arr = JSON.parse(raw);
    return coerceArray(arr);
  } catch {
    // fallback if someone stuffed a CSV in there
    return coerceArray(bookEl.dataset.completedChapters);
  }
}

function setCompletedOnEl(bookEl, arr) {
  const uniqSorted = Array.from(new Set(arr.map(Number)))
    .filter(n => Number.isFinite(n) && n > 0)
    .sort((a, b) => a - b);
  bookEl.dataset.completedChapters = JSON.stringify(uniqSorted);
  applyCompletedUI(bookEl, uniqSorted);
}

// -----------------------------------------------

function drawCourse(course) {
  const box = document.createElement('div');
  box.classList.add('course-box', 'collapsed');
  box.dataset.courseId = course.id;
  box.setAttribute('aria-expanded', 'false'); // default closed

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
  } else {
    // only show delete when there are NO books
    const delBtn = deleteCourseButton(course.id, box);
    box.appendChild(delBtn);
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
  li.dataset.numChapters = String(book.numChapters);
  // store as JSON array string for stability
  li.dataset.completedChapters = JSON.stringify(book.completedChapters ?? []);

  const title = document.createElement('div');
  title.classList.add('book-title');
  title.textContent = book.title;

  const meta = document.createElement('div');
  meta.classList.add('book-meta');
  // keep existing copy; you can change to "X • Y chapters" later if desired
  meta.textContent = `${book.author} • ${book.numChapters} chapters`;

  const chapters = drawChapters(book);

  const more = document.createElement('button');
  more.type = 'button';
  more.className = 'book-menu-btn';
  more.textContent = '⋯';              // U+22EF
  more.addEventListener('click', (e) => {
    e.stopPropagation();               // don’t collapse the course
    handleBookMenuClick(book.id, more);
  });

  li.append(title, meta, chapters, more);

  return li;
}

function drawChapters(book) {
  const wrap = document.createElement('div');
  wrap.classList.add('chapters');

  const completed = new Set(coerceArray(book.completedChapters));

  for (let i = 1; i <= book.numChapters; i++) {
    const ch = document.createElement('div');
    ch.classList.add('chapter-box');
    ch.textContent = i;
    if (completed.has(i)) ch.classList.add('completed');

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

// Update UI classes based on completed array
function applyCompletedUI(bookEl, completedArr) {
  const completed = new Set(coerceArray(completedArr));
  const boxes = bookEl.querySelectorAll('.chapter-box');
  boxes.forEach((box) => {
    const n = Number(box.textContent);
    if (completed.has(n)) box.classList.add('completed');
    else box.classList.remove('completed');
  });
}

async function handleChapterClick(bookEl, bookId, n) {
  const before = getCompletedFromEl(bookEl);
  const has = before.includes(n);
  const optimistic = has ? before.filter(x => x !== n) : [...before, n];

  // optimistic UI
  setCompletedOnEl(bookEl, optimistic);

  try {
    const after = has
      ? await bookService.removeProgress(bookId, n)
      : await bookService.addProgress(bookId, n);

    // server returns []int; trust it
    setCompletedOnEl(bookEl, after ?? optimistic);
  } catch (err) {
    // rollback
    setCompletedOnEl(bookEl, before);
    console.error('Failed to toggle chapter', err);
  }
}

function handleAddBook(courseId) {
  openBookForm(courseId, {
    onSubmit: async (data, { close }) => {
      try {
        const numChapters = Number(data.numChapters);
        const payload = {
          courseId,
          title: data.title,
          author: data.author,
          numChapters,
          ...(data.link && data.link.trim() ? { link: data.link.trim() } : {}),
        };
        const created = await bookService.create(payload);
        // ensure completedChapters is an array for new items
        created.completedChapters = created.completedChapters ?? [];
        appendBookToUI(courseId, created);
        close();
      } catch (err) {
        console.error('Create failed:', err);
      }
    }
  });
}

function appendBookToUI(courseId, book) {
  const box = document.querySelector(`.course-box[data-course-id="${courseId}"]`);
  if (!box) return;

  // remove course delete button if this was an empty course
  const delBtn = box.querySelector('.course-delete-btn');
  if (delBtn) delBtn.remove();

  // find or create the books wrapper + list
  let booksWrap = box.querySelector('.books');
  let list;
  if (!booksWrap) {
    booksWrap = document.createElement('div');
    booksWrap.classList.add('books');

    list = document.createElement('ul');
    list.classList.add('books-list');
    booksWrap.appendChild(list);

    const addBtn = box.querySelector('.add-book-btn');
    if (addBtn) box.insertBefore(booksWrap, addBtn);
    else box.appendChild(booksWrap);
  } else {
    list = booksWrap.querySelector('.books-list');
  }

  list.appendChild(drawBook(book));
}

function handleBookMenuClick(bookId, btnEl) {
  openBookMenu({
    bookId,
    anchorEl: btnEl,
    onDelete: async (id) => {
      try {
        await bookService.delete(id);

        const row = btnEl.closest('.book-item');
        if (row) {
          const list = row.parentElement; // .books-list
          row.remove();

          if (list && list.children.length === 0) {
            const courseBox = list.closest('.course-box');

            // remove the empty books wrapper
            const wrap = list.closest('.books');
            if (wrap) wrap.remove();

            // now that there are no books, ensure the course delete button is present
            if (courseBox && !courseBox.querySelector('.course-delete-btn')) {
              const courseId = courseBox.dataset.courseId;
              const delBtn = deleteCourseButton(courseId, courseBox);
              courseBox.appendChild(delBtn);
            }
          }
        }
        return true;
      } catch (e) {
        console.error('Delete failed:', e);
        return false;
      }
    }
  });
}

function addCourseBtn() {
  const footer = document.createElement('div');
  footer.className = 'course-footer';

  const addCourseBtn = document.createElement('button');
  addCourseBtn.type = 'button';
  addCourseBtn.className = 'add-book-btn add-course-btn';
  addCourseBtn.textContent = '＋ Add course';
  addCourseBtn.addEventListener('click', () => handleAddCourse());

  footer.appendChild(addCourseBtn);
  container.appendChild(footer);
}

function deleteCourseButton(courseId, boxEl) {
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'add-book-btn course-delete-btn';
  btn.textContent = 'Delete course';
  btn.addEventListener('click', (e) => {
    e.stopPropagation();              // don’t toggle collapse
    handleDeleteCourse(courseId, boxEl);
  });
  return btn;
}

function handleAddCourse() {
  openCourseForm({
    onSubmit: async (data, { close }) => {
      try {
        const payload = {
          year: Number(data.year),
          term: Number(data.term),
          code: data.code,
          name: data.name,
        };
        const created = await courseService.create(payload);
        appendCourseToUI(created);
        close();
      } catch (err) {
        console.error('Create course failed:', err);
      }
    }
  });
}

async function handleDeleteCourse(courseId, boxEl) {
  if (!confirm("Delete this course?")) return;

  // optimistic remove
  const parent = boxEl.parentElement;
  const nextSibling = boxEl.nextSibling;
  boxEl.remove();

  try {
    await courseService.delete(courseId);
  } catch (err) {
    console.error("Delete course failed:", err);
    // rollback on failure
    if (nextSibling) parent.insertBefore(boxEl, nextSibling);
    else parent.appendChild(boxEl);
    alert(err.message || "Delete failed");
  }
}

function appendCourseToUI(course) {
  const box = drawCourse({ ...course, books: course.books || [] });
  const footer = container.querySelector('.course-footer');
  if (footer) container.insertBefore(box, footer);
  else container.appendChild(box);
}
