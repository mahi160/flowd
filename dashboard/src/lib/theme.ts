export type Theme = "dark" | "light" | "system";
export const THEMES: Theme[] = ["dark", "light", "system"];

export function resolveTheme(t: Theme): "dark" | "light" {
  if (t === "system") {
    return matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  return t;
}

export function loadTheme(): Theme {
  try {
    return (localStorage.getItem("fw-theme") as Theme) || "system";
  } catch {
    return "system";
  }
}

export function saveTheme(t: Theme): void {
  try {
    localStorage.setItem("fw-theme", t);
  } catch {}
}

export function cycleTheme(current: Theme): Theme {
  return THEMES[(THEMES.indexOf(current) + 1) % THEMES.length];
}
