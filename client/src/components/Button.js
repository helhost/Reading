export default function Button({
  label,
  type = "default", // "default" | "primary" | "success" | "warn" | "danger"
  onClick,
  disabled = false,
}) {
  const btn = document.createElement("button");
  btn.className = `btn btn--${type}`;
  btn.textContent = label;
  if (disabled) btn.disabled = true;
  if (onClick) btn.addEventListener("click", onClick);
  return btn;
}
