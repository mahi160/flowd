import { el } from '../utils.js';

export function Summary(DATA) {
  const { focus, code, byProject, byCommand, topRepo } = DATA;
  const card = el('div', 'card summary-card');

  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Summary'),
    el('div', 'card-sub', `${DATA.period} · narrative`),
  ));

  const top3 = arr => arr.slice(0, 3).map(x => `${x.name} ${x.min}m`).join(' · ') || '—';

  function row(term, ...ddChildren) {
    const dd = el('dd'); dd.append(...ddChildren);
    const d  = el('div', null, el('dt', null, term), dd);
    return d;
  }

  const dl = el('dl', 'summary-list',
    row('Focus', el('b', 'tnum', focus.totalMin + 'm'), ` across `, el('b', 'tnum', focus.blocks), ` blocks`),
    byProject.length ? row('Projects', top3(byProject)) : null,
    byCommand.length ? row('Tools',    top3(byCommand))  : null,
    row('Code', el('span', 'tnum', focus.totalMin + 'm'),
      ` (`, el('span', 'tnum diff-add', `+${code.added}`), ` `, el('span', 'tnum diff-rm', `−${code.removed}`), `)`),
  );

  if (topRepo.name && topRepo.name !== '—') {
    const dd = el('dd', null, el('code', null, topRepo.name));
    if (topRepo.branch) dd.append(' on ', el('code', null, topRepo.branch));
    dl.appendChild(el('div', null, el('dt', null, 'Top repo'), dd));
  }

  card.appendChild(el('div', 'summary-scroll', dl));
  return card;
}
