import type { RawPayload, Item, TimelineEntry, DayBucket, ToolSummary } from "./types";

// PALETTE removed - components use their own COLORS
// // Components (Donut, Languages) maintain their own COLORS arrays.
// The Item.color field is set here for API compatibility but not used for rendering.
const COLORS = ["#6366f1","#f59e0b","#ef4444","#10b981","#8b5cf6","#ec4899","#14b8a6","#f97316"];

function list(obj?: Record<string, number>): Item[] {
  return Object.entries(obj || {})
    .filter(([, v]) => v > 0)
    .sort((a, b) => b[1] - a[1])
    .map(([name, min], i) => ({ name, min, color: COLORS[i % COLORS.length] }));
}

export function transform(raw: RawPayload = {}) {
  const heat = raw.heatmap || [];
  const isToday = raw.period !== "week";
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
    period: raw.period === "week" ? ("week" as const) : ("today" as const),
    generated: raw.generated || "",
    machine: raw.machine || "—",
    os: raw.os || "",
    topRepo: { name: raw.top_repo || "—", branch: raw.top_branch || "" },
    focus: {
      totalMin: raw.total_focus_min || 0,
      blocks: raw.total_blocks || 0,
      switches: raw.total_switches || 0,
    },
    code: {
      files: raw.files_changed || 0,
      added: raw.lines_added || 0,
      removed: raw.lines_removed || 0,
    },
    byProject: list(raw.by_project),
    byCommand: list(raw.by_tool),
    byLanguage: list(raw.languages),
    hourly,
    weekHourlyByDay: byDayHour,
    weekDays: Object.entries(byDay).map(([label, min]): DayBucket => {
      const [day, date = ""] = label.split(" ");
      return { day, date, min };
    }),
    streakDays,
    // 30-cell grid: each cell = one day, newest at index 29.
    // v=0 no activity, v=3 part of current streak, v=1 dim (streak recently broken).
    // We only know the streak COUNT from the backend (not per-day minutes for 30d),
    // so we mark streak days accurately and leave the rest dark.
    streakCells: Array.from({ length: 30 }, (_, i) => {
      const ago = 29 - i; // 0 = today, 29 = 29 days ago
      return {
        d: i,
        v: ago < streakDays ? (ago === 0 ? 4 : 3) : 0,
      };
    }),
    timeline: (raw.timeline || []).map((b): TimelineEntry => ({
      from: b.start,
      to: b.end,
      project: b.repo || null,
      branch: b.branch || null,
      focus: b.focus || 0,
      switches: b.switches || 0,
      ai: b.ai || null,
    })),
    aiRecap: raw.ai_recap || null,
    aiPerBlock: raw.ai_per_block || 0,
    aiTools: (raw.ai_tools || []) as ToolSummary[],
    hasData: (raw.total_blocks || 0) > 0 || (raw.ai_tools || []).length > 0,
  };
}

export type Data = ReturnType<typeof transform>;
