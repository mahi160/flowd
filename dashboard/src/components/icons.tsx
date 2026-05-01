import { Item } from "../types";
import { PALETTE } from "../lib/data";

export const ICONS = {
  dark: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>`,
  light: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="4" stroke="currentColor" stroke-width="1.6"/><path d="M12 3v2M12 19v2M3 12h2M19 12h2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4 7 17M17 7l1.4-1.4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>`,
  system: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="13" rx="2" stroke="currentColor" stroke-width="1.6"/><path d="M8 21h8M12 17v4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>`,
};

export const SvgIcon = ({ html, size = 16 }: { html: string; size?: number }) => (
  <span dangerouslySetInnerHTML={{ __html: html }} style={{ width: size, height: size, display: "inline-flex" }} />
);

export const BranchIcon = ({ size = 12 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none">
    <path d="M4 2v8a3 3 0 0 0 3 3h5M12 13l-2-2m2 2-2 2M4 6h5a3 3 0 0 0 3-3V2" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

export const FlowMark = ({ size = 40 }: { size?: number }) => {
  const id = `flowmark-grad-${size}`;
  return (
    <svg width={size} height={size} viewBox="0 0 64 64" fill="none" aria-hidden="true">
      <defs>
        <linearGradient id={id} x1="10" y1="54" x2="54" y2="10">
          <stop stopColor="var(--moss)" />
          <stop offset=".55" stopColor="var(--fern)" />
          <stop offset="1" stopColor="var(--amber)" />
        </linearGradient>
      </defs>
      <path d="M17 41c9 0 10-18 22-18 5 0 8 3 8 8 0 12-15 18-25 9" stroke={`url(#${id})`} strokeWidth="7" strokeLinecap="round" />
      <circle cx="17" cy="41" r="5" fill="var(--moss)" />
      <circle cx="47" cy="31" r="5" fill="var(--amber)" />
    </svg>
  );
};
