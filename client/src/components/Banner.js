let _bannerRoot = null;

export function mountBanner({ user, onLogout, links = [] } = {}) {
  if (_bannerRoot) _bannerRoot.remove();

  const banner = document.createElement("header");
  banner.className = "banner";

  // Left: user or brand info
  const left = document.createElement("div");
  left.className = "banner__left";
  left.textContent = user?.email || "Not signed in";

  // Center: optional navigation links
  const center = document.createElement("nav");
  center.className = "banner__center";

  if (Array.isArray(links) && links.length) {
    for (const link of links) {
      const a = document.createElement("a");
      a.textContent = link.label;
      a.href = link.href;
      if (link.navigo) a.setAttribute("data-navigo", "");
      a.className = "banner__link";
      center.appendChild(a);
    }
  }

  // Right: action buttons (logout / login)
  const right = document.createElement("div");
  right.className = "banner__right";

  if (user?.email) {
    const logoutBtn = document.createElement("button");
    logoutBtn.type = "button";
    logoutBtn.className = "bf-btn";
    logoutBtn.textContent = "Log out";
    logoutBtn.addEventListener("click", async () => {
      try {
        await onLogout?.();
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

  // assemble
  banner.append(left, center, right);
  document.body.prepend(banner);

  _bannerRoot = banner;
  return () => unmountBanner();
}

/**
 * Updates the banner’s displayed user or links dynamically.
 */
export function updateBanner({ user, links = [] } = {}) {
  if (!_bannerRoot) return;

  const left = _bannerRoot.querySelector(".banner__left");
  const center = _bannerRoot.querySelector(".banner__center");
  const right = _bannerRoot.querySelector(".banner__right");

  // update left
  if (left) left.textContent = user?.email || "Not signed in";

  // update center
  if (center) {
    center.innerHTML = "";
    for (const link of links) {
      const a = document.createElement("a");
      a.textContent = link.label;
      a.href = link.href;
      if (link.navigo) a.setAttribute("data-navigo", "");
      a.className = "banner__link";
      center.appendChild(a);
    }
  }

  // update right
  if (right) {
    right.innerHTML = "";
    if (user?.email) {
      const logoutBtn = document.createElement("button");
      logoutBtn.type = "button";
      logoutBtn.className = "bf-btn";
      logoutBtn.textContent = "Log out";
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
 * Removes the banner from DOM.
 */
export function unmountBanner() {
  if (_bannerRoot) {
    _bannerRoot.remove();
    _bannerRoot = null;
  }
}
