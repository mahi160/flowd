import { el } from '../utils.js';

function HeatLegend() {
  const leg = el('span', 'heat-legend', 'less ');
  for (const v of [0.05, 0.3, 0.55, 0.8, 1]) {
    const c = el('span', 'heat-cell');
    c.style.setProperty('--i', v);
    leg.appendChild(c);
  }
  leg.appendChild(document.createTextNode(' more'));
  return leg;
}

function heatCell(intensity, tooltip) {
  const c = el('span', 'heat-cell');
  c.style.setProperty('--i', intensity);
  c.title = tooltip;
  return c;
}

export function ActivityHeatmap(DATA) {
  const card = el('div', 'card heatmap-card');
  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Activity heatmap'),
    el('div', 'card-sub', 'today · focus blocks by start time'),
  ));

  const axis = el('div', 'heat-axis');
  for (const l of ['8a', '10a', '12p', '2p', '4p', '6p', '8p']) axis.appendChild(el('span', null, l));
  card.appendChild(axis);

  const cells = el('div', 'heat-cells');
  let peakIntensity = 0, peakHour = 8;
  for (let i = 0; i < 60; i++) {
    const rawHour = 8 + Math.floor(i * 12 / 60);
    const v = DATA.hourly[Math.min(23, rawHour)] || 0;
    const noise = ((i * 13) % 7) / 6;
    const intensity = Math.max(0, Math.min(1, (v / 45) * (0.5 + noise * 0.6)));
    if (intensity > peakIntensity) { peakIntensity = intensity; peakHour = rawHour; }
    cells.appendChild(heatCell(intensity, `~${rawHour}:00 — ≈${Math.round(intensity * 30)}m`));
  }
  card.appendChild(cells);

  const foot = el('div', 'heat-foot',
    HeatLegend(),
    el('span', 'heat-peak', peakIntensity > 0 ? `peak ~${peakHour}:00` : ''),
  );
  card.appendChild(foot);
  return card;
}

export function WeekHeatmap(DATA) {
  const card = el('div', 'card heatmap-card');
  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Activity heatmap'),
    el('div', 'card-sub', 'last 7 days × hour of day'),
  ));

  const days = Object.entries(DATA.weekHourlyByDay);
  const maxV = Math.max(1, ...days.flatMap(([, hrs]) => hrs));
  const grid = el('div', 'week-heat-grid');

  for (const [day, hrs] of days) {
    const cells = el('div', 'week-heat-cells');
    hrs.forEach((v, h) => cells.appendChild(heatCell(v / maxV, `${day} ${String(h).padStart(2, '0')}:00 · ${v}m`)));
    grid.appendChild(el('div', 'week-heat-row', el('span', 'week-heat-lbl', day), cells));
  }
  card.appendChild(grid);

  card.appendChild(el('div', 'heat-foot', HeatLegend(), el('span', null, '24h × 7 days')));
  return card;
}
