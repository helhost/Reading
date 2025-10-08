import { Home } from "../Icons/Home.js";

let _bannerRoot = null;
let _callbacks = { onLogout: null, onHomeClick: null };

export function mountBanner({ user, onLogout, onHomeClick } = {}) {
  if (!onHomeClick) throw new Error("onHomeClick is required");
  if (_bannerRoot) _bannerRoot.remove();
  _callbacks = { onLogout: onLogout ?? null, onHomeClick };

  const banner = document.createElement("header");
  banner.className = "banner";

  // Home: mandatory house icon button
  const homeBtn = document.createElement("button");
  homeBtn.type = "button";
  homeBtn.className = "banner__home";
  homeBtn.setAttribute("aria-label", "Home");
  homeBtn.innerHTML = Home;
  homeBtn.addEventListener("click", () => {
    try {
      _callbacks.onHomeClick?.();
    } catch (e) {
      console.error("Home click handler failed:", e);
    }
  });

  // Left: user/brand info
  const left = document.createElement("div");
  left.className = "banner__left";
  left.textContent = user?.email || "Not signed in";

  // Right: actions
  const right = document.createElement("div");
  right.className = "banner__right";
  if (user?.email) {
    const logoutBtn = document.createElement("button");
    logoutBtn.type = "button";
    logoutBtn.className = "bf-btn";
    logoutBtn.textContent = "Log out";
    logoutBtn.addEventListener("click", async () => {
      try {
        await _callbacks.onLogout?.();
      } catch (e) {
        console.error("Logout failed:", e);
      }
    });
    right.appendChild(logoutBtn);
  } else {
    const loginLink = document.createElement("a");
    loginLink.href = "/#/login";
    loginLink.textContent = "Log in";
    loginLink.className = "banner__link";
    right.appendChild(loginLink);
  }

  // assemble (home, left, right)
  banner.append(homeBtn, left, right);
  document.body.prepend(banner);

  const setBannerH = () => {
    const h = banner.getBoundingClientRect().height; // includes borders
    document.documentElement.style.setProperty("--banner-h", `${h}px`);
  };

  // set once after layout paints
  requestAnimationFrame(setBannerH);

  // keep it in sync on resize/content changes
  window.addEventListener("resize", setBannerH);
  new ResizeObserver(setBannerH).observe(banner);
  _bannerRoot = banner;
  return () => unmountBanner();
}

/**
 * Update displayed user state. (No center/links anymore)
 */
export function updateBanner({ user, onLogout } = {}) {
  if (!_bannerRoot) return;
  if (onLogout !== undefined) _callbacks.onLogout = onLogout;

  const left = _bannerRoot.querySelector(".banner__left");
  const right = _bannerRoot.querySelector(".banner__right");

  // left
  if (left) left.textContent = user?.email || "Not signed in";

  // right
  if (right) {
    right.innerHTML = "";
    if (user?.email) {
      const logoutBtn = document.createElement("button");
      logoutBtn.type = "button";
      logoutBtn.className = "bf-btn";
      logoutBtn.textContent = "Log out";
      logoutBtn.addEventListener("click", async () => {
        try {
          await _callbacks.onLogout?.();
        } catch (e) {
          console.error("Logout failed:", e);
        }
      });
      right.appendChild(logoutBtn);
    } else {
      const loginLink = document.createElement("a");
      loginLink.href = "/#/login";
      loginLink.textContent = "Log in";
      loginLink.className = "banner__link";
      right.appendChild(loginLink);
    }
  }
}

/**
 * Remove banner from DOM.
 */
export function unmountBanner() {
  if (_bannerRoot) {
    _bannerRoot.remove();
    _bannerRoot = null;
  }
}
