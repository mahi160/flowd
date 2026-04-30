import { el } from '../utils.js';

function fmtH(h) { return h === 0 ? '12a' : h < 12 ? `${h}a` : h === 12 ? '12p' : `${h - 12}p`; }

export function HourPattern(DATA) {
  const data = DATA.hourly;
  const max  = Math.max(...data, 1);
  const card = el('div', 'card hour-card');

  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Best hours'),
    el('div', 'card-sub', 'focus by hour of day'),
  ));

  const bars = el('div', 'hour-bars');
  data.forEach((v, i) => {
    const bar  = el('div', 'hour-bar');
    const fill = el('span', 'hour-fill');
    fill.style.setProperty('--h', `${(v / max) * 100}%`);
    bar.appendChild(fill);
    if (i % 4 === 0) bar.appendChild(el('span', 'hour-tick', fmtH(i)));
    bars.appendChild(bar);
  });
  card.appendChild(bars);

  const peakH = data.indexOf(Math.max(...data));
  const first = data.findIndex(v => v > 0);
  const last  = data.reduce((a, v, i) => v > 0 ? i : a, -1);

  function stat(label, val) {
    return el('div', null, el('span', 'eyebrow', label), el('span', 'tnum', val));
  }
  const stats = el('div', 'hour-stats', stat('peak', fmtH(peakH)));
  if (first >= 0) stats.appendChild(stat('start', fmtH(first)));
  if (last  >= 0) stats.appendChild(stat('end',   fmtH(last)));
  card.appendChild(stats);
  return card;
}
