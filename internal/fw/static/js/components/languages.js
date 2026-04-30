import { el, fmtHM } from '../utils.js';

export function Languages(DATA) {
  const langs = DATA.byLanguage.slice(0, 8);
  const card  = el('div', 'card lang-card');

  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Languages'),
    el('div', 'card-sub', 'by time in session'),
  ));

  if (!langs.length) {
    card.appendChild(el('p', null, 'No language data — inferred from git diffs.'));
    return card;
  }

  // Stacked colour bar
  const bar = el('div', 'lang-bar');
  langs.forEach(l => {
    const seg = el('span', 'lang-seg');
    seg.style.flex = l.min;
    seg.style.background = l.color;
    seg.title = `${l.name} · ${l.min}m`;
    bar.appendChild(seg);
  });
  card.appendChild(bar);

  // List rows
  const list = el('ul', 'lang-list');
  const topMin = langs[0].min;
  langs.forEach(l => {
    const swatch = el('span', 'legend-swatch'); swatch.style.background = l.color;
    const fill   = el('span'); fill.style.width = `${(l.min / topMin) * 100}%`; fill.style.background = l.color;
    const miniBar = el('div', 'lang-bar-mini', fill);
    list.appendChild(el('li', null, swatch, el('span', 'lang-name', l.name), miniBar, el('span', 'tnum lang-val', l.min + 'm')));
  });
  card.appendChild(list);

  const total = langs.reduce((a, b) => a + b.min, 0);
  const foot  = el('div', 'lang-foot', el('span', 'eyebrow', 'total'));
  const tot   = el('span', 'tnum lang-total font-display', fmtHM(total) + ' ');
  tot.appendChild(el('span', 'dim', 'tracked'));
  foot.appendChild(tot);
  card.appendChild(foot);
  return card;
}
