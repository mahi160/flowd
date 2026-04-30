import { el, FlowMark } from '../utils.js';
import { cycle, getCurrent, getIcon } from '../theme.js';

export function Header(DATA, view, onViewChange) {
  const hdr = el('header', 'header');

  // ── Brand ──────────────────────────────────────────────────
  const brand = el('div', 'brand');
  const brandText = el('div');

  const sub = el('div', 'brand-sub',
    el('span', 'eyebrow', 'Generated'),
    ' ',
    el('span', 'brand-stamp tnum', DATA.generated),
  );
  if (DATA.hasData) {
    const chip = el('span', 'chip', el('span', 'status-dot'), ' ' + DATA.period);
    sub.append(' · ', chip);
  }

  brandText.append(el('div', 'brand-name font-display', 'flowd'), sub);
  brand.append(FlowMark(40), brandText);

  // ── Controls ───────────────────────────────────────────────
  const right = el('div', 'header-right');

  const seg = el('div', 'seg');
  for (const v of ['today', 'week']) {
    const btn = el('button', 'seg-btn' + (view === v ? ' is-active' : ''), v === 'today' ? 'Today' : 'Week');
    if (v !== DATA.period) btn.title = `Run \`fw dashboard ${v}\` to generate ${v} data`;
    btn.addEventListener('click', () => onViewChange(v));
    seg.appendChild(btn);
  }

  const themeBtn = el('button', 'icon-btn');
  themeBtn.id = 'theme-btn';
  themeBtn.innerHTML = getIcon(getCurrent());
  themeBtn.title = `Theme: ${getCurrent()}`;
  themeBtn.addEventListener('click', () => cycle());

  right.append(seg, themeBtn);
  hdr.append(brand, right);
  return hdr;
}
