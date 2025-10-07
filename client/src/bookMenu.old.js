let _openBookMenuEl = null;

export function openBookMenu({ bookId, anchorEl, onDelete }) {
  closeBookMenu();

  const menu = document.createElement('div');
  menu.className = 'book-popover';
  menu.addEventListener('click', e => e.stopPropagation());

  const list = document.createElement('div');
  list.className = 'book-popover-list';

  const del = document.createElement('button');
  del.type = 'button';
  del.className = 'book-popover-item';
  del.textContent = 'Delete book';
  del.addEventListener('click', async e => {
    e.stopPropagation();
    if (!confirm("Are you sure you want to delete this book?")) return;

    try {
      del.disabled = true;
      del.textContent = 'Deleting…';
      const ok = typeof onDelete === 'function' ? await onDelete(bookId) : true;
      if (ok) closeBookMenu();
      else {
        del.disabled = false;
        del.textContent = 'Delete book';
      }
    } catch (err) {
      console.error('Delete failed:', err);
      del.disabled = false;
      del.textContent = 'Delete book';
    }
  });

  list.appendChild(del);
  menu.appendChild(list);
  document.body.appendChild(menu);

  const r = anchorEl.getBoundingClientRect();
  const m = menu.getBoundingClientRect();
  const top = window.scrollY + r.bottom + 6;
  let left = window.scrollX + r.right - m.width;
  const minLeft = window.scrollX + 8;
  if (left < minLeft) left = minLeft;
  menu.style.top = `${top}px`;
  menu.style.left = `${left}px`;

  const onDocClick = e => { if (!menu.contains(e.target) && e.target !== anchorEl) closeBookMenu(); };
  const onKey = e => { if (e.key === 'Escape') closeBookMenu(); };

  document.addEventListener('click', onDocClick, { capture: true, once: true });
  document.addEventListener('keydown', onKey, { once: true });

  _openBookMenuEl = menu;
}

export function closeBookMenu() {
  if (_openBookMenuEl) {
    _openBookMenuEl.remove();
    _openBookMenuEl = null;
  }
}
