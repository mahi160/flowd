import { el, BranchIcon } from '../utils.js';

export function Timeline(DATA) {
  const entries = [...DATA.timeline].reverse();
  const maxF    = Math.max(1, ...entries.map(e => e.focus));
  const card    = el('div', 'card timeline-card');

  card.appendChild(el('div', 'card-head',
    el('div', 'card-title', 'Timeline'),
    el('div', 'card-sub', 'focus blocks · newest first'),
  ));

  if (!entries.length) {
    card.appendChild(el('p', null, 'No blocks recorded yet.'));
    return card;
  }

  const list = el('ol', 'timeline');

  entries.forEach((entry, i) => {
    const isIdle = !entry.project;
    const row    = el('li', 'tl-row' + (isIdle ? ' tl-idle' : ''));

    // Time column
    row.appendChild(el('div', 'tl-time tnum', `${entry.from} → ${entry.to}`));

    // Rail (dot + connecting line)
    const dot = el('span', 'tl-dot');
    dot.style.setProperty('--c', isIdle ? 'var(--ink-5)' : 'var(--accent)');
    const rail = el('div', 'tl-rail', dot);
    if (i < entries.length - 1) rail.appendChild(el('span', 'tl-line'));
    row.appendChild(rail);

    // Body
    const body = el('div', 'tl-body');
    if (isIdle) {
      body.appendChild(el('div', 'tl-head', el('span', 'dim', '— idle')));
    } else {
      const head = el('div', 'tl-head', el('span', 'tl-project', entry.project));
      if (entry.branch) head.appendChild(el('span', 'branch-tag', BranchIcon(10), ' ' + entry.branch));
      body.appendChild(head);

      const meta = el('div', 'tl-meta', el('span', 'tnum', entry.focus + 'm'), ' focus');
      if (entry.switches > 0) meta.append(el('span', 'dim', ' · '), el('span', 'tnum', entry.switches), ' context switches');
      body.appendChild(meta);

      const barFill = el('span'); barFill.style.width = `${(entry.focus / maxF) * 100}%`;
      body.appendChild(el('div', 'tl-bar', barFill));

      if (entry.ai) body.appendChild(el('div', 'tl-ai', entry.ai));
    }
    row.appendChild(body);
    list.appendChild(row);
  });

  card.appendChild(el('div', 'timeline-scroll', list));
  return card;
}
