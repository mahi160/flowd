import { el } from '../utils.js';

export function StreakCard(DATA) {
  const s    = DATA.streakDays;
  const card = el('div', 'card streak-card');

  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Streak'),
    el('div', 'card-sub', 'last 30 days'),
  ));

  // Hero number
  const sub = s === 0 ? 'start coding today' : s >= 14 ? 'on fire 🔥' : s >= 7 ? 'keep it up 🌿' : 'building momentum';
  const heroRight = el('div', 'streak-hero-right',
    el('span', 'streak-hero-label', 'day streak'),
    el('span', 'streak-hero-sub eyebrow', sub),
  );
  card.appendChild(el('div', 'streak-hero', el('span', 'streak-big font-display tnum', s), heroRight));

  // Launchpad grid — 6 cols × 5 rows = 30 pads
  const grid = el('div', 'streak-grid');
  DATA.streakCells.forEach(cell => {
    const intensity = cell.v / 4;
    const pad = el('span', 'streak-pad');

    pad.style.background = `color-mix(in oklch, var(--moss) ${intensity * 85}%, var(--bg-inset))`;
    pad.style.border      = `1px solid color-mix(in oklch, var(--moss) ${intensity * 40}%, var(--hairline-soft))`;

    if (intensity > 0.15) {
      const sz = 5 + intensity * 9, sp = intensity * 2.5;
      pad.style.boxShadow = [
        `inset 0 1px 0 rgba(255,255,255,${(intensity * 0.12).toFixed(2)})`,
        `0 0 ${sz.toFixed(1)}px ${sp.toFixed(1)}px color-mix(in oklch, var(--moss) ${Math.round(intensity * 55)}%, transparent)`,
      ].join(', ');
    }

    if (cell.d === 29) pad.style.outline = '2px solid color-mix(in oklch, var(--moss) 70%, transparent)';
    pad.title = cell.d === 29 ? 'today' : `${29 - cell.d}d ago`;
    grid.appendChild(pad);
  });
  card.appendChild(grid);
  card.appendChild(el('div', 'streak-axis', el('span', null, '30d ago'), el('span', null, 'today')));
  return card;
}
