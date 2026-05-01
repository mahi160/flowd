export type Theme = 'dark' | 'light' | 'system';
export const THEMES: Theme[] = ['dark', 'light', 'system'];

export function resolveTheme(t: Theme): 'dark' | 'light' {
  return t === 'system'
    ? matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    : t;
}

export function applyTheme(theme: Theme): void {
  document.documentElement.dataset.theme = resolveTheme(theme);
  localStorage.setItem('fw-theme', theme);
}

export function savedTheme(): Theme {
  return (localStorage.getItem('fw-theme') as Theme) ?? 'system';
}
