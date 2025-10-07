export default function Card({ title, actionLabel, onAction, onClick } = {}) {
  const card = document.createElement("div");
  card.className = "card";

  const header = document.createElement("div");
  header.className = "card-header";

  const h3 = document.createElement("h3");
  h3.className = "card-title";
  h3.textContent = title;

  header.appendChild(h3);

  if (actionLabel && onAction) {
    const btn = document.createElement("button");
    btn.className = "card-action-btn";
    btn.textContent = actionLabel;
    btn.addEventListener("click", (e) => {
      e.stopPropagation(); // prevent firing onClick
      onAction(e);
    });
    header.appendChild(btn);
  }

  card.appendChild(header);

  if (onClick) {
    card.classList.add("card--clickable");
    card.addEventListener("click", onClick);
  }

  return card;
}
