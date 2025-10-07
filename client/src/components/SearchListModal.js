import openModal from "./Modal.js";
import MiniCard from "./MiniCard.js";

// items: array<any>
// getTitle(item): string
// onPick(item): void (modal auto-closes after onPick runs)
// actionLabel: string (button label on each mini-card)
export function openSearchListModal({
  title = "Select item",
  items = [],
  getTitle = (x) => String(x),
  onPick,
  actionLabel = "Select",
} = {}) {
  const m = openModal({ title });

  const wrap = document.createElement("div");
  wrap.className = "searchlist";

  // search input
  const input = document.createElement("input");
  input.type = "search";
  input.className = "searchlist__input";
  input.placeholder = "Search…";

  // list container
  const list = document.createElement("div");
  list.className = "searchlist__list";

  wrap.append(input, list);
  m.setBody(wrap);

  // render helper
  function render(filtered) {
    list.innerHTML = "";
    if (!filtered.length) {
      const empty = document.createElement("div");
      empty.className = "searchlist__empty";
      empty.textContent = "No matches.";
      list.appendChild(empty);
      return;
    }
    for (const item of filtered) {
      const card = MiniCard({
        title: getTitle(item),
        actionLabel,
        onAction: () => {
          try { onPick?.(item); }
          finally { m.close(); }
        },
      });
      list.appendChild(card);
    }
  }

  // initial render (unsorted minimalism)
  render(items);

  // simple debounced filter
  let t = null;
  input.addEventListener("input", () => {
    clearTimeout(t);
    const q = input.value.trim().toLowerCase();
    t = setTimeout(() => {
      if (!q) return render(items);
      const filtered = items.filter((it) => getTitle(it).toLowerCase().includes(q));
      render(filtered);
    }, 120);
  });

  // focus
  setTimeout(() => input.focus(), 0);

  return m; // caller can also close programmatically if needed
}

export default openSearchListModal;
