export default function MiniCard({ title, actionLabel, onAction }) {
  const row = document.createElement("div");
  row.className = "mini-card";

  const h = document.createElement("div");
  h.className = "mini-card__title";
  h.textContent = title;

  row.appendChild(h);

  if (actionLabel && onAction) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "mini-card__btn";
    btn.textContent = actionLabel;
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      onAction();
    });
    row.appendChild(btn);
  }

  return row;
}
