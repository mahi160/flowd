// ── DOM / SVG helpers ────────────────────────────────────────
const SVG_NS = 'http://www.w3.org/2000/svg';

/** Create an HTML element. children can be string | Node | Array. */
export function el(tag, cls, ...children) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  for (const c of children) mount(e, c);
  return e;
}

/** Create an SVG element with given attributes. */
export function svgEl(tag, attrs = {}) {
  const e = document.createElementNS(SVG_NS, tag);
  for (const [k, v] of Object.entries(attrs)) e.setAttribute(k, v);
  return e;
}

/** Append a child — handles strings, numbers, Nodes, arrays, null. */
export function mount(parent, child) {
  if (child == null) return;
  if (Array.isArray(child))        { child.forEach(c => mount(parent, c)); return; }
  if (child instanceof Node)       { parent.appendChild(child); return; }
  if (typeof child !== 'object')   { parent.appendChild(document.createTextNode(String(child))); return; }
}

/** Format minutes as "2h 30m". */
export function fmtHM(m) {
  const h = Math.floor(m / 60), mm = m % 60;
  return h === 0 ? `${mm}m` : mm === 0 ? `${h}h` : `${h}h ${mm}m`;
}

/** Chart colour palette using CSS custom properties. */
export const PALETTE = [
  'var(--c1)', 'var(--c2)', 'var(--c3)', 'var(--c4)',
  'var(--c5)', 'var(--c6)', 'var(--c7)', 'var(--c8)',
];

/** SVG donut arc path (rO = outer radius, rI = inner). */
export function arcPath(cx, cy, rO, rI, a0, a1) {
  const large = a1 - a0 > Math.PI ? 1 : 0;
  const pt = (r, a) => [cx + r * Math.cos(a), cy + r * Math.sin(a)];
  const [x0, y0] = pt(rO, a0), [x1, y1] = pt(rO, a1);
  const [xi0, yi0] = pt(rI, a1), [xi1, yi1] = pt(rI, a0);
  return `M${x0} ${y0} A${rO} ${rO} 0 ${large} 1 ${x1} ${y1} `
       + `L${xi0} ${yi0} A${rI} ${rI} 0 ${large} 0 ${xi1} ${yi1}Z`;
}

/** Inline SVG icon for a git branch. */
export function BranchIcon(size = 11) {
  const s = svgEl('svg', { width: size, height: size, viewBox: '0 0 16 16', fill: 'none' });
  s.innerHTML = '<path d="M5 3v10M11 3a2 2 0 1 0-2 2v3c0 1.5-1.5 2-3 2-1 0-1.5.3-2 1" stroke="currentColor" stroke-width="1.4"/>'
              + '<circle cx="5" cy="3" r="1.5" stroke="currentColor" stroke-width="1.4"/>';
  return s;
}

/** flowd logo mark SVG. */
export function FlowMark(size = 40) {
  const s = svgEl('svg', { width: size, height: size, viewBox: '0 0 40 40', 'aria-hidden': 'true' });
  s.innerHTML = `<defs>
    <linearGradient id="fmg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="var(--moss)"/>
      <stop offset="1" stop-color="var(--amber)"/>
    </linearGradient>
  </defs>
  <rect x="1" y="1" width="38" height="38" rx="11" fill="url(#fmg)" opacity=".18"/>
  <rect x="1" y="1" width="38" height="38" rx="11" fill="none" stroke="url(#fmg)" stroke-width="1.3"/>
  <path d="M11 28c4-1 7-4 8-8 1-4 4-7 9-8" stroke="var(--moss)" stroke-width="2" fill="none" stroke-linecap="round"/>
  <circle cx="28" cy="12" r="2.2" fill="var(--amber)"/>`;
  return s;
}

/** Inline sparkline SVG from an array of values. */
export function Sparkline(data) {
  const W = 220, H = 26, max = Math.max(...data, 1);
  const step = W / (data.length - 1);
  const pts = data.map((v, i) => `${i * step},${H - (v / max) * H}`).join(' L ');
  const path = `M ${pts}`;
  const s = svgEl('svg', { width: '100%', height: H, viewBox: `0 0 ${W} ${H}`, preserveAspectRatio: 'none', class: 'sparkline' });
  s.innerHTML = `<defs>
    <linearGradient id="sp" x1="0" x2="0" y1="0" y2="1">
      <stop offset="0" stop-color="var(--accent)" stop-opacity=".4"/>
      <stop offset="1" stop-color="var(--accent)" stop-opacity="0"/>
    </linearGradient>
  </defs>
  <path d="${path} L ${W},${H} L 0,${H} Z" fill="url(#sp)"/>
  <path d="${path}" stroke="var(--accent)" stroke-width="1.6" fill="none" stroke-linecap="round" stroke-linejoin="round"/>`;
  return s;
}
