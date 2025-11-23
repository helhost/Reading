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

// "YYYY-MM-DD" + hour -> unix seconds at LOCAL hour
function dateHourToUnix(dateStr, hourStr) {
  if (!dateStr) return null;
  const [y, m, d] = dateStr.split("-").map((x) => Number(x));
  if (!y || !m || !d) return null;

  let h = Number(hourStr);
  if (!Number.isFinite(h)) h = 0;
  if (h < 0) h = 0;
  if (h > 23) h = 23;

  const dt = new Date(y, m - 1, d, h, 0, 0, 0); // local at given hour
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

  // Row: date + hour input
  const rowInput = document.createElement("div");
  rowInput.className = "dp-row";

  // date only
  const input = document.createElement("input");
  input.type = "date";
  input.className = "dp-input";

  // hour only
  const hourInput = document.createElement("input");
  hourInput.type = "number";
  hourInput.className = "dp-input";
  hourInput.min = 0;
  hourInput.max = 23;
  hourInput.placeholder = "hh";
  hourInput.addEventListener("input", () => {
    let val = hourInput.value.replace(/\D/g, ""); // keep digits only
    if (val === "") return;

    if (val.length > 2) val = val.slice(0, 2);    // at most 2 digits
    let num = Number(val);
    if (num > 23) num = 23;
    if (num < 0) num = 0;

    hourInput.value = String(num);
  });

  // initial value for date and hour
  const initialDate = toDate(initial);
  if (initialDate) {
    input.value = toLocalInputDateString(initialDate);
    hourInput.value = String(initialDate.getHours());
  }

  // min and max for date
  const minDate = toDate(min);
  if (minDate) input.min = toLocalInputDateString(minDate);

  const maxDate = toDate(max);
  if (maxDate) input.max = toLocalInputDateString(maxDate);

  rowInput.append(input, hourInput);

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
      const ts = dateHourToUnix(input.value, hourInput.value);
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
