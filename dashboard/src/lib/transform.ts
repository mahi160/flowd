import type { RawPayload, RawPeriodData, Item, TimelineEntry, DayBucket, ToolSummary, CalDay, MonthBar } from "./types";

// Consistent colours used across all chart components.
export const CHART_COLORS = ["#6366f1","#f59e0b","#ef4444","#10b981","#8b5cf6","#ec4899","#14b8a6","#f97316"];

function list(obj?: Record<string, number>): Item[] {
  return Object.entries(obj || {})
    .filter(([, v]) => v > 0)
    .sort((a, b) => b[1] - a[1])
    .map(([name, min], i) => ({ name, min, color: CHART_COLORS[i % CHART_COLORS.length] }));
}

// ── Main transform ──────────────────────────────────────────────────────────
// Accepts the full raw payload and the currently selected period.
// Returns a typed Data object for that period; shared fields (machine, streak
// etc.) are lifted from the top level.
export function transform(raw: RawPayload = {}, period: string = "today") {
  const p: RawPeriodData = raw.periods?.[period] ?? {};
  const heat = p.heatmap || [];
  const isToday = period === "today";

  const hourly = Array(24).fill(0) as number[];
  const byDay: Record<string, number> = {};
  const byDayHour: Record<string, number[]> = {};

  for (const c of heat) {
    const h = isToday ? Math.floor(c.hour / 2) : c.hour;
    if (h >= 0 && h < 24) hourly[h] += c.minute;
    byDay[c.day] = (byDay[c.day] || 0) + c.minute;
    byDayHour[c.day] ||= Array(24).fill(0);
    if (c.hour >= 0 && c.hour < 24) byDayHour[c.day][c.hour] += c.minute;
  }

  const streakDays = raw.streak_days || 0;

  return {
    // Shared across all periods
    generated:    raw.generated || "",
    machine:      raw.machine   || "—",
    os:           raw.os        || "",
    streakDays,
    aiRecap:      raw.ai_recap  || null,

    // Period identity
    period: period as "today" | "week" | "month" | "year" | "all",

    // Period-specific counts
    focus: {
      totalMin: p.total_focus_min || 0,
      blocks:   p.total_blocks    || 0,
      switches: p.total_switches  || 0,
    },
    code: {
      files:   p.files_changed || 0,
      added:   p.lines_added   || 0,
      removed: p.lines_removed || 0,
    },
    topRepo:   { name: p.top_repo || "—", branch: p.top_branch || "" },
    byProject: list(p.by_project),
    byCommand: list(p.by_tool),
    byLanguage: list(p.languages),

    // Heatmap data (today = 48 half-hour buckets, week = 7×24)
    hourly,
    weekHourlyByDay: byDayHour,
    weekDays: Object.entries(byDay).map(([label, min]): DayBucket => {
      const [day, date = ""] = label.split(" ");
      return { day, date, min };
    }),

    // Streak cells
    streakCells: Array.from({ length: 30 }, (_, i) => {
      const ago = 29 - i;
      return { d: i, v: ago < streakDays ? (ago === 0 ? 4 : 3) : 0 };
    }),

    // Timeline
    timeline: (p.timeline || []).map((b): TimelineEntry => ({
      from:     b.start,
      to:       b.end,
      project:  b.repo    || null,
      branch:   b.branch  || null,
      focus:    b.focus   || 0,
      switches: b.switches || 0,
      ai:       b.ai      || null,
    })),

    // AI sessions
    aiTools: (p.ai_tools || []) as ToolSummary[],

    // Calendar / all-time heatmap data
    calDays:    (p.cal_days    || []) as CalDay[],
    monthBars:  (p.month_bars  || []) as MonthBar[],

    // Derived period stats
    trackingSince: p.tracking_since || "",
    activeDays:    p.active_days    || 0,
    bestDayDate:   p.best_day_date  || "",
    bestDayMin:    p.best_day_min   || 0,

    // hasData: period exists in the payload and has some activity
    hasData: !!(raw.periods?.[period] &&
      ((p.total_blocks || 0) > 0 || (p.ai_tools || []).length > 0)),

    // anyData: at least one period has data (for the overall empty state)
    anyData: Object.values(raw.periods || {}).some(
      (pd) => (pd.total_blocks || 0) > 0 || (pd.ai_tools || []).length > 0
    ),
  };
}

export type Data = ReturnType<typeof transform>;
