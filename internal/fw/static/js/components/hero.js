import { el, fmtHM, Sparkline, BranchIcon } from '../utils.js';

export function HeroStrip(DATA, periodLabel) {
  const strip = el('div', 'hero-strip');
  strip.append(FocusCard(DATA, periodLabel), MachineCard(DATA), RepoCard(DATA), CodeCard(DATA));
  return strip;
}

function FocusCard(DATA, periodLabel) {
  const { focus } = DATA;
  const card = el('div', 'hero-card hero-primary card');

  const num = el('div', 'hero-num font-display tnum');
  num.append(
    String(Math.floor(focus.totalMin / 60)), el('span', 'hero-unit', 'h'),
    String(focus.totalMin % 60),             el('span', 'hero-unit', 'm'),
  );

  const sub = el('div', 'hero-sub');
  sub.append(
    el('span', 'tnum', focus.blocks), ' focus blocks',
    el('span', 'dim', ' · '),
    el('span', 'tnum', focus.switches), ' context switches',
  );

  card.append(el('div', 'eyebrow', `Focus ${periodLabel}`), num, sub, Sparkline(DATA.hourly));
  return card;
}

function MachineCard(DATA) {
  const card = el('div', 'hero-card card');
  const barFill = el('span'); barFill.style.width = '50%';
  const bar = el('div', 'machine-bar', barFill);
  card.append(el('div', 'eyebrow', 'Machine'), el('div', 'hero-mid font-display', DATA.machine), el('div', 'hero-sub', DATA.os), bar);
  return card;
}

function RepoCard(DATA) {
  const { topRepo, byProject } = DATA;
  const card = el('div', 'hero-card card');
  card.append(el('div', 'eyebrow', 'Top repo'), el('div', 'hero-mid font-display', topRepo.name));
  if (topRepo.branch) {
    const tag = el('span', 'branch-tag', BranchIcon(), ' ' + topRepo.branch);
    card.appendChild(el('div', 'hero-sub', tag));
  }
  if (byProject[0]) card.appendChild(el('div', 'hero-mini tnum', `${byProject[0].name} · ${byProject[0].min}m`));
  return card;
}

function CodeCard(DATA) {
  const { code } = DATA;
  const card = el('div', 'hero-card card');
  const sub = el('div', 'hero-sub');
  sub.append(el('span', 'tnum diff-add', `+${code.added}`), el('span', 'dim', ' / '), el('span', 'tnum diff-rm', `−${code.removed}`));
  card.append(el('div', 'eyebrow', 'Code'), el('div', 'hero-mid font-display tnum', `${code.files} files`), sub);
  return card;
}
