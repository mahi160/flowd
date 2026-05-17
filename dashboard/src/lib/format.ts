export function fmtHM(m = 0): string {
  m = Math.max(0, m);
  const h = Math.floor(m / 60);
  const r = m % 60;
  return h ? `${h}h ${r}m` : `${r}m`;
}

export function fmtHour(h: number): string {
  return h === 0 ? "12a" : h < 12 ? `${h}a` : h === 12 ? "12p" : `${h - 12}p`;
}

export function fmtTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return `${n}`;
}

export function fmtCost(n: number): string {
  if (n >= 1) return `$${n.toFixed(2)}`;
  if (n > 0) return `$${n.toFixed(4)}`;
  return "—";
}

const polar = (cx: number, cy: number, r: number, a: number): [number, number] => [
  cx + r * Math.cos(a),
  cy + r * Math.sin(a),
];

export function arcPath(
  cx: number, cy: number,
  ro: number, ri: number,
  start: number, end: number,
): string {
  const large = end - start >= Math.PI ? 1 : 0;
  const [x1, y1] = polar(cx, cy, ro, start);
  const [x2, y2] = polar(cx, cy, ro, end);
  const [x3, y3] = polar(cx, cy, ri, end);
  const [x4, y4] = polar(cx, cy, ri, start);
  return `M ${x1} ${y1} A ${ro} ${ro} 0 ${large} 1 ${x2} ${y2} L ${x3} ${y3} A ${ri} ${ri} 0 ${large} 0 ${x4} ${y4} Z`;
}
