import { PALETTE } from './utils.js';

function objToArr(obj, vk = 'min') {
  return Object.entries(obj || {})
    .filter(([, v]) => v > 0)
    .sort((a, b) => b[1] - a[1])
    .map(([name, value], i) => ({ name, [vk]: value, color: PALETTE[i % PALETTE.length] }));
}

function computeHourly(heatmap) {
  const h = new Array(24).fill(0);
  for (const c of (heatmap || [])) {
    const i = Math.floor(c.hour / 2);
    if (i >= 0 && i < 24) h[i] += c.minute;
  }
  return h;
}

function computeWeekHourly(heatmap) {
  const byDay = {};
  for (const c of (heatmap || [])) {
    if (!byDay[c.day]) byDay[c.day] = new Array(24).fill(0);
    if (c.hour >= 0 && c.hour < 24) byDay[c.day][c.hour] += c.minute;
  }
  return byDay;
}

function computeWeekDays(heatmap) {
  const byDay = {};
  for (const c of (heatmap || [])) byDay[c.day] = (byDay[c.day] || 0) + c.minute;
  return Object.entries(byDay).map(([lbl, min]) => {
    const [day, date] = lbl.split(' ');
    return { day, date: date || '', min };
  });
}

function makeStreakCells(streakDays) {
  return Array.from({ length: 30 }, (_, i) => {
    const ago = 29 - i;
    if (ago < streakDays) return { d: i, v: ago === 0 ? 4 : Math.min(4, ((i * 7 + 3) % 3) + 2) };
    return { d: i, v: (ago < streakDays + 8 && (i * 13) % 4 === 0) ? 1 : 0 };
  });
}

export function transform(RAW) {
  const isToday   = RAW.period === 'today';
  const byProject = objToArr(RAW.by_project);
  const byCommand = objToArr(RAW.by_tool);
  const byLanguage = objToArr(RAW.languages, 'min');

  const hourly = isToday
    ? computeHourly(RAW.heatmap)
    : (() => {
        const h = new Array(24).fill(0);
        for (const c of (RAW.heatmap || [])) if (c.hour >= 0 && c.hour < 24) h[c.hour] += c.minute;
        return h;
      })();

  return {
    period:          RAW.period || 'today',
    generated:       RAW.generated || '',
    machine:         RAW.machine || '—',
    os:              RAW.os || '',
    topRepo:         { name: RAW.top_repo || '—', branch: RAW.top_branch || '' },
    focus:           { totalMin: RAW.total_focus_min || 0, blocks: RAW.total_blocks || 0, switches: RAW.total_switches || 0 },
    code:            { files: RAW.files_changed || 0, added: RAW.lines_added || 0, removed: RAW.lines_removed || 0 },
    byProject, byCommand, byLanguage, hourly,
    weekDays:        computeWeekDays(RAW.heatmap),
    weekHourlyByDay: computeWeekHourly(RAW.heatmap),
    streakCells:     makeStreakCells(RAW.streak_days || 0),
    streakDays:      RAW.streak_days || 0,
    timeline:        (RAW.timeline || []).map(b => ({
      from: b.start, to: b.end,
      project: b.repo || null, branch: b.branch || null,
      focus: b.focus, switches: b.switches, ai: b.ai || null,
      intensity: Math.min(1, (b.focus || 0) / 30),
    })),
    aiRecap:    RAW.ai_recap   || null,
    aiPerBlock: RAW.ai_per_block || 0,
    hasData:    (RAW.total_blocks || 0) > 0,
  };
}
