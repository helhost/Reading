import openModal from "./Modal.js";
import Button from "./Button.js";
import CalendarSvc from "../services/calendar.js";
import { Calendar as CalendarSVG } from "../Icons/Calendar.js";
import { Copy as CopySVG } from "../Icons/Copy.js";

/* ---------- helpers ---------- */
function svgFromString(svgStr) {
  const doc = new DOMParser().parseFromString(svgStr, "image/svg+xml");
  return doc.documentElement; // <svg>
}
function normalizeIconColors(svgEl) {
  // use currentColor so CSS controls light/dark
  const all = [svgEl, ...svgEl.querySelectorAll("*")];
  for (const el of all) {
    el.removeAttribute("style");
    if (el.hasAttribute("fill")) el.setAttribute("fill", "currentColor");
    if (el.hasAttribute("stroke")) el.setAttribute("stroke", "currentColor");
  }
  svgEl.setAttribute("fill", "currentColor");
}

export function mountCalendarSubscribe() {
  // avoid duplicates across route changes
  if (document.getElementById("calendar-subscribe-fab")) return;

  // FAB (icon-only)
  const fab = document.createElement("button");
  fab.id = "calendar-subscribe-fab";
  fab.type = "button";
  fab.className = "fab fab--bl fab--round fab--calendar";
  fab.setAttribute("aria-label", "Subscribe to calendar");

  const calIcon = svgFromString(CalendarSVG);
  normalizeIconColors(calIcon);
  calIcon.classList.add("icon");
  fab.appendChild(calIcon);

  document.body.appendChild(fab);

  fab.addEventListener("click", async () => {
    const modal = openModal({ title: "Subscribe to your calendar" });

    // ===== build body (no innerHTML) =====
    const body = document.createElement("div");

    const label = document.createElement("label");
    label.className = "cal-sub__label";
    label.setAttribute("for", "cal-sub-url");
    label.textContent = "Subscription URL";

    // input + Copy inline
    const row = document.createElement("div");
    row.className = "cal-sub__row";

    const input = document.createElement("input");
    input.id = "cal-sub-url";
    input.className = "input-field cal-sub__input";
    input.readOnly = true;
    input.value = "Loading…";

    const copyBtn = Button({ label: "", type: "default" });
    copyBtn.classList.add("btn--icon");
    const copyIcon = svgFromString(CopySVG);
    normalizeIconColors(copyIcon);
    copyIcon.classList.add("icon");
    const copyText = document.createElement("span");
    copyText.textContent = "Copy";
    copyBtn.append(copyIcon, copyText);

    row.append(input, copyBtn);

    // actions row: New URL button
    const actions = document.createElement("div");
    actions.className = "cal-sub__actions";
    const rotateBtn = Button({ label: "New URL", type: "warn" });
    actions.append(rotateBtn);

    body.append(label, row, actions);
    modal.setBody(body);

    // ===== data wiring =====
    let tokenData;
    try {
      tokenData = await CalendarSvc.getToken(); // { token, urlPath }
      if (!tokenData) throw new Error("not logged in");
    } catch {
      input.value = "Sign in to generate a link.";
      copyBtn.disabled = true;
      rotateBtn.disabled = true;
      return;
    }

    const setUrl = (urlPath) => {
      input.value = CalendarSvc.toAbsoluteUrl(urlPath);
    };
    setUrl(tokenData.urlPath);

    // Copy
    copyBtn.addEventListener("click", async () => {
      try {
        await navigator.clipboard.writeText(input.value);
        copyText.textContent = "Copied";
        setTimeout(() => (copyText.textContent = "Copy"), 1100);
      } catch {
        copyText.textContent = "Copy failed";
        setTimeout(() => (copyText.textContent = "Copy"), 1100);
      }
    });

    // Rotate
    rotateBtn.addEventListener("click", async () => {
      rotateBtn.disabled = true;
      const prev = rotateBtn.textContent;
      rotateBtn.textContent = "Generating…";
      try {
        const next = await CalendarSvc.rotateToken(); // { token, urlPath }
        setUrl(next.urlPath);
        rotateBtn.textContent = "New URL";
      } catch {
        rotateBtn.textContent = "Failed. Try again";
        setTimeout(() => (rotateBtn.textContent = prev), 1100);
      } finally {
        rotateBtn.disabled = false;
      }
    });
  });
}
