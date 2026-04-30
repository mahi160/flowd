import { el, fmtHM } from '../utils.js';

export function WeekBars(DATA) {
  const days = DATA.weekDays;
  const maxM = Math.max(1, ...days.map(d => d.min));
  const card = el('div', 'card week-card');

  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Week at a glance'),
    el('div', 'card-sub', 'focus per day'),
  ));

  const bars = el('div', 'week-bars');
  days.forEach(d => {
    const fill = el('span', 'week-fill');
    fill.style.setProperty('--h', `${(d.min / maxM) * 100}%`);
    if (d.min > 0) fill.appendChild(el('span', 'week-fill-val tnum', fmtHM(d.min)));

    const track = el('div', 'week-track', fill);
    const label = el('div', 'week-label', el('div', 'week-day', d.day), el('div', 'week-date tnum', d.date));
    bars.appendChild(el('div', 'week-bar', track, label));
  });

  card.appendChild(bars);
  return card;
}
