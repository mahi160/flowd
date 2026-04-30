const THEMES = ["dark", "light", "system"];

const ICONS = {
  dark: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none">
    <path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z"
      stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/>
  </svg>`,
  light: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="12" r="4" stroke="currentColor" stroke-width="1.6"/>
    <path d="M12 3v2M12 19v2M3 12h2M19 12h2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4 7 17M17 7l1.4-1.4"
      stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>
  </svg>`,
  system: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none">
    <rect x="2" y="4" width="20" height="13" rx="2" stroke="currentColor" stroke-width="1.6"/>
    <path d="M8 21h8M12 17v4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>
  </svg>`,
};

let current = localStorage.getItem("fw-theme") || "system";

function resolved() {
  return current === "system"
    ? window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light"
    : current;
}

export function apply(t) {
  current = t;
  localStorage.setItem("fw-theme", t);
  document.documentElement.setAttribute("data-theme", resolved());
  const btn = document.getElementById("theme-btn");
  if (btn) {
    btn.innerHTML = ICONS[t];
    btn.title = `Theme: ${t}`;
  }
}

export function cycle() {
  apply(THEMES[(THEMES.indexOf(current) + 1) % THEMES.length]);
}

export function getCurrent() {
  return current;
}

export function getIcon(t) {
  return ICONS[t];
}

export function initTheme() {
  apply(current);
  window
    .matchMedia("(prefers-color-scheme: dark)")
    .addEventListener("change", () => {
      if (current === "system") apply("system");
    });
}
