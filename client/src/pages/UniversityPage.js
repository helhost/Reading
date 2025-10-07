import getMe from "../util/index.js";
import universityService from "../services/universities.js";
import Card from "../components/Card.js";

export default async function UniversityPage() {
  const me = await getMe();
  if (!me) return;

  const root = ensureRoot();
  root.innerHTML = "";

  // Load memberships + all unis so we can show names
  let myUniIds = new Set();
  try {
    const [myUnis, allUnis] = await Promise.all([
      universityService.getMyUniversities(), // [{ universityId }]
      universityService.getAll(),            // [{ id, name }]
    ]);

    // id -> name map
    const nameById = new Map(allUnis.map(u => [u.id, u.name]));

    // render joined cards
    const list = document.createElement("div");
    list.className = "course-container"; // reuse your width/breakpoint style
    for (const m of myUnis) {
      const id = m.universityId;
      const card = Card({
        title: nameById.get(id) || id,
        actionLabel: "Leave",
        onAction: () => leaveUni(id),
      });
      list.appendChild(card);
    }
    root.appendChild(list);

    myUniIds = new Set(myUnis.map(m => m.universityId));
    console.log("My universities:", myUnis);
  } catch (err) {
    console.error("Failed to fetch user universities:", err);
  }

  // Button: fetch all and log the remaining ones
  const btn = document.createElement("button");
  btn.textContent = "Get all universities";
  btn.className = "add-book-btn";
  btn.addEventListener("click", async () => {
    const remaining = await getRemainingUnis(myUniIds);
    console.log("All universities (NOT joined):", remaining);
  });

  root.appendChild(btn);
}

// fetch all, exclude those I already joined
async function getRemainingUnis(myIds) {
  try {
    const all = await universityService.getAll();
    return all.filter(u => !myIds.has(u.id));
  } catch (err) {
    console.error("Failed to fetch all universities:", err);
    return [];
  }
}

async function leaveUni(id) {
  console.log("leaving uni", id);
  // TODO: call DELETE /api/user-universities with { universityId: id } when you implement it
}

function ensureRoot() {
  let node = document.getElementById("app");
  if (!node) {
    node = document.createElement("main");
    node.id = "app";
    document.body.appendChild(node);
  }
  return node;
}
