import { el, svgEl, fmtHM, arcPath } from '../utils.js';

export function Donut(title, subtitle, items, vk = 'min') {
  const card = el('div', 'card donut-card');
  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', title),
    subtitle ? el('div', 'card-sub', subtitle) : null,
  ));

  if (!items.length) {
    card.appendChild(el('div', null, 'No data'));
    return card;
  }

  const total = items.reduce((a, b) => a + (b[vk] || 0), 0);
  const W = 220, H = 220, cx = 110, cy = 110, rO = 86, rI = 56;

  // Build arc segments
  let acc = -Math.PI / 2;
  const segs = items.map(item => {
    const v = item[vk] || 0;
    const frac = total ? v / total : 0;
    const s = acc, e = acc + frac * Math.PI * 2;
    acc = e;
    return { ...item, v, frac, s, e };
  });

  // SVG
  const s = svgEl('svg', { viewBox: `0 0 ${W} ${H}`, width: '100%', height: '100%', class: 'donut' });
  const paths = segs.map(seg => {
    const p = svgEl('path', { d: arcPath(cx, cy, rO, rI, seg.s, seg.e), fill: seg.color });
    p.style.cursor = 'default';
    s.appendChild(p);
    return p;
  });
  s.append(
    svgEl('circle', { cx, cy, r: rI - 1, fill: 'none', stroke: 'var(--hairline-soft)', 'stroke-width': '1' }),
    svgEl('circle', { cx, cy, r: rO + 1, fill: 'none', stroke: 'var(--hairline-soft)', 'stroke-width': '1' }),
  );

  // Center text (switches between total and hovered segment)
  const cTotal = el('div', 'donut-total font-display tnum', fmtHM(total));
  const cCap   = el('div', 'donut-cap eyebrow', 'total');
  const cPct   = el('div', 'donut-pct font-display tnum');
  const cName  = el('div', 'donut-name');
  const cVal   = el('div', 'donut-val tnum');
  [cPct, cName, cVal].forEach(e => { e.style.display = 'none'; });

  const center = el('div', 'donut-center', cTotal, cCap, cPct, cName, cVal);
  const wrap   = el('div', 'donut-wrap', s, center);
  card.appendChild(wrap);

  function onEnter(i) {
    paths.forEach((p, j) => { p.style.opacity = j === i ? '1' : '0.35'; });
    cTotal.style.display = 'none'; cCap.style.display = 'none';
    [cPct, cName, cVal].forEach(e => { e.style.display = ''; });
    cPct.textContent  = Math.round(segs[i].frac * 100) + '%';
    cName.textContent = segs[i].name;
    cVal.textContent  = segs[i].v + 'm';
  }
  function onLeave() {
    paths.forEach(p => { p.style.opacity = '1'; });
    cTotal.style.display = ''; cCap.style.display = '';
    [cPct, cName, cVal].forEach(e => { e.style.display = 'none'; });
  }

  paths.forEach((p, i) => {
    p.addEventListener('mouseenter', () => onEnter(i));
    p.addEventListener('mouseleave', onLeave);
  });

  // Legend
  const leg = el('ul', 'legend');
  segs.forEach((seg, i) => {
    const swatch = el('span', 'legend-swatch'); swatch.style.background = seg.color;
    const row = el('li', 'legend-row',
      swatch,
      el('span', 'legend-name', seg.name),
      el('span', 'legend-val tnum', seg.v + 'm'),
      el('span', 'legend-pct tnum', Math.round(seg.frac * 100) + '%'),
    );
    row.addEventListener('mouseenter', () => onEnter(i));
    row.addEventListener('mouseleave', onLeave);
    leg.appendChild(row);
  });
  card.appendChild(leg);
  return card;
}
