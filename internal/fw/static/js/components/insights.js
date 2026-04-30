import { el } from '../utils.js';

function InsightIcon() {
  const icon = el('div', 'insights-icon');
  icon.innerHTML = `<svg width="15" height="15" viewBox="0 0 24 24" fill="none">
    <path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.5 5.5l2 2M16.5 16.5l2 2M5.5 18.5l2-2M16.5 7.5l2-2"
      stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>
    <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.6"/>
  </svg>`;
  return icon;
}

function insightItem(tag, tagCls, content) {
  const item = el('div', 'insight-item', el('span', `insight-tag ${tagCls}`, tag));
  const p = el('p');
  if (typeof content === 'string') p.textContent = content;
  else p.innerHTML = content; // trusted internal HTML only
  item.appendChild(p);
  return item;
}

export function Insights(DATA) {
  const card = el('div', 'card insights-card');

  if (DATA.aiRecap) {
    const chip = el('span', 'chip', 'recap'); chip.style.marginLeft = 'auto';
    card.append(el('div', 'insights-head', InsightIcon(), el('div', 'card-title', 'AI recap'), chip));
    card.appendChild(insightItem('summary', 'good', DATA.aiRecap));
    return card;
  }

  if (DATA.aiPerBlock > 0) {
    const n = DATA.aiPerBlock;
    const chip = el('span', 'chip', `${n} block${n === 1 ? '' : 's'}`); chip.style.marginLeft = 'auto';
    card.append(el('div', 'insights-head', InsightIcon(), el('div', 'card-title', 'AI insights'), chip));
    card.appendChild(insightItem('inline', '', 'Per-block AI summaries are inline in the timeline. Run <code>fw dashboard --ai-recap</code> for an aggregate.'));
    return card;
  }

  card.append(el('div', 'insights-head', InsightIcon(), el('div', 'card-title', 'AI insights')));
  card.appendChild(insightItem('setup', '', 'Set <code>ai_enabled: true</code> and <code>ai_command</code> in your config to see AI insights here.'));
  return card;
}
