export default function Card({ title, actionLabel, onAction }) {
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
    btn.addEventListener("click", onAction);
    header.appendChild(btn);
  }

  card.appendChild(header);
  return card;
}
