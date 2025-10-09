import openModal from "./Modal.js";
import Button from "./Button.js";

function toDate(val) {
  if (val == null) return null;
  if (val instanceof Date) return val;
  if (Number.isFinite(val)) return new Date(val * 1000);
  return null;
}

function pad2(n) { return n < 10 ? `0${n}` : `${n}`; }

// Format Date -> "YYYY-MM-DD" (LOCAL)
function toLocalInputDateString(d) {
  if (!d) return "";
  const y = d.getFullYear();
  const m = pad2(d.getMonth() + 1);
  const day = pad2(d.getDate());
  return `${y}-${m}-${day}`;
}

// "YYYY-MM-DD" -> unix seconds at LOCAL midnight
function dateInputToUnix(dateStr) {
  if (!dateStr) return null;
  const [y, m, d] = dateStr.split("-").map((x) => Number(x));
  if (!y || !m || !d) return null;
  const dt = new Date(y, m - 1, d, 0, 0, 0, 0); // local midnight
  return Math.floor(dt.getTime() / 1000);
}

export default function openDatePicker({
  anchorEl,
  initial = null,
  min = null,
  max = null,
  onPick,
  onClose,
  centered = false,        // center on screen when true (ignore anchor)
} = {}) {
  if (!centered && !anchorEl) {
    throw new Error("DatePicker requires anchorEl (unless centered=true)");
  }

  const modal = centered
    ? openModal({
      centered: true,
      overlayClass: "modal-overlay--centered",
      cardClass: "datepicker-popover modal-card--centered",
      onClose,
    })
    : openModal({
      anchorEl,
      placement: "bottom-start",
      offset: 8,
      cardClass: "menu-popover datepicker-popover",
      onClose,
    });

  const body = document.createElement("div");
  body.className = "dp-body";
  body.setAttribute("role", centered ? "dialog" : "group");
  if (centered) body.setAttribute("aria-label", "Pick a date");

  // Row: date input
  const rowInput = document.createElement("div");
  rowInput.className = "dp-row";

  const input = document.createElement("input");
  input.type = "date";
  input.className = "dp-input";

  const initialDate = toDate(initial);
  if (initialDate) input.value = toLocalInputDateString(initialDate);

  const minDate = toDate(min);
  if (minDate) input.min = toLocalInputDateString(minDate);

  const maxDate = toDate(max);
  if (maxDate) input.max = toLocalInputDateString(maxDate);

  rowInput.appendChild(input);

  // Row: quick chips
  const rowQuick = document.createElement("div");
  rowQuick.className = "dp-row dp-quick";

  function setInputTo(dateObj) {
    input.value = toLocalInputDateString(dateObj);
  }

  const todayChip = document.createElement("button");
  todayChip.type = "button";
  todayChip.className = "dp-chip";
  todayChip.textContent = "Today";
  todayChip.addEventListener("click", () => setInputTo(new Date()));

  const nextWeekChip = document.createElement("button");
  nextWeekChip.type = "button";
  nextWeekChip.className = "dp-chip";
  nextWeekChip.textContent = "Next week";
  nextWeekChip.addEventListener("click", () => {
    const d = new Date();
    d.setDate(d.getDate() + 7);
    setInputTo(d);
  });

  rowQuick.append(todayChip, nextWeekChip);

  // Row: actions
  const rowActions = document.createElement("div");
  rowActions.className = "dp-row dp-actions";

  const clearBtn = Button({
    label: "Clear",
    type: "warn",
    onClick: () => {
      try { onPick?.(null); } finally { modal.close(); }
    },
  });

  const saveBtn = Button({
    label: "Save",
    type: "primary",
    onClick: () => {
      const ts = dateInputToUnix(input.value);
      try { onPick?.(ts); } finally { modal.close(); }
    },
  });

  rowActions.append(clearBtn, saveBtn);

  body.append(rowInput, rowQuick, rowActions);
  modal.setBody(body);

  // Keyboard niceties: Enter = Save, Esc handled by modal, focus input
  const onKey = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      saveBtn?.click();
    }
  };
  body.addEventListener("keydown", onKey);
  modal.onCleanup?.(() => body.removeEventListener("keydown", onKey));

  // Focus input for faster entry
  setTimeout(() => input.focus(), 0);

  return modal;
}
