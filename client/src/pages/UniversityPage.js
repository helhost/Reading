import { getMe } from "../util/index.js";
import universityService from "../services/universities.js";
import Card from "../components/Card.js";
import { OpenSearchListModal } from "../components/SearchListModal.js";
import { Toast } from "../components/Toast.js";

import Button from "../components/Button.js";

export default async function UniversityPage() {
  const me = await getMe();
  if (!me) return;

  const root = ensureRoot();
  root.innerHTML = "";

  // containers
  const list = document.createElement("div");
  list.className = "uni-container";

  const footer = document.createElement("div");
  footer.className = "uni-footer";

  const addBtn = Button({
    label: "＋ Add university",
    type: "primary",
    onClick: () => {
      const remaining = allUnis.filter(u => !myIds.has(u.id));
      if (!remaining.length) {
        Toast("info", "No more universities to join.");
        return;
      }
      OpenSearchListModal({
        title: "Join a university",
        items: remaining,
        getTitle: (u) => u.name,
        actionLabel: "Join",
        onPick: async (u) => {
          try {
            await universityService.join(u.id);
            Toast("success", `Joined ${u.name}`);
            await refresh();
          } catch (e) {
            Toast("error", e?.message || "Failed to join");
          }
        },
      });
    },
  });


  footer.appendChild(addBtn);
  root.append(list, footer);

  // state
  let allUnis = [];          // [{ id, name }]
  let myMemberships = [];    // [{ universityId }]
  let myIds = new Set();     // Set<string>

  // initial load
  await refresh();

  // leave flow
  async function leaveUni(id, name) {
    try {
      await universityService.leave(id);
      Toast("success", `Left ${name}`);
      await refresh(); // update UI after success
    } catch (e) {
      Toast("error", e?.message || "Failed to leave");
    }
  }

  // fetch + render
  async function refresh() {
    try {
      const [my, all] = await Promise.all([
        universityService.getMyUniversities(),
        universityService.getAll(),
      ]);
      myMemberships = my;
      allUnis = all;
      myIds = new Set(myMemberships.map(m => m.universityId));
      renderJoined();
    } catch (err) {
      console.error(err);
      Toast("error", err?.message || "Failed to load universities");
    }
  }

  function renderJoined() {
    list.innerHTML = "";

    const nameById = new Map(allUnis.map(u => [u.id, u.name]));

    if (!myMemberships.length) {
      const empty = document.createElement("div");
      empty.className = "card";
      empty.textContent = "You haven't joined any universities yet.";
      list.appendChild(empty);
      return;
    }

    for (const m of myMemberships) {
      const id = m.universityId;
      const name = nameById.get(id) || id;

      const card = Card({
        title: name,
        actionLabel: "Leave",
        onAction: (e) => {
          // prevent any parent click handlers later
          e?.stopPropagation?.();
          leaveUni(id, name);
        },
      });

      list.appendChild(card);
    }
  }
}

/* tiny helper */
function ensureRoot() {
  let node = document.getElementById("app");
  if (!node) {
    node = document.createElement("main");
    node.id = "app";
    document.body.appendChild(node);
  }
  return node;
}
