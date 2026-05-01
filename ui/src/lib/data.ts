import type { RawPayload, ParsedData, Item } from "./types";
import { PALETTE } from "./palette";

const list = (obj?: Record<string, number>): Item[] =>
  Object.entries(obj || {})
    .filter(([, v]) => v > 0)
    .sort((a, b) => b[1] - a[1])
    .map(([name, min], i) => ({
      name,
      min,
      color: PALETTE[i % PALETTE.length],
    }));

export function transform(raw: Partial<RawPayload> = {}): ParsedData {
  const heat = raw.heatmap || [];
  const isToday = raw.period !== "week";
  const hourly = Array<number>(24).fill(0);
  const byDay: Record<string, number> = {};
  const byDayHour: Record<string, number[]> = {};

  for (const c of heat) {
    const h = isToday ? Math.floor(c.hour / 2) : c.hour;
    if (h >= 0 && h < 24) hourly[h] += c.minute;
    byDay[c.day] = (byDay[c.day] || 0) + c.minute;
    byDayHour[c.day] ??= Array<number>(24).fill(0);
    if (c.hour >= 0 && c.hour < 24) byDayHour[c.day][c.hour] += c.minute;
  }

  const streak = raw.streak_days || 0;

  return {
    period: raw.period === "week" ? "week" : "today",
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
    weekDays: Object.entries(byDay).map(([label, min]) => {
      const [day, date = ""] = label.split(" ");
      return { day, date, min };
    }),
    streakDays: streak,
    streakCells: Array.from({ length: 30 }, (_, i) => {
      const ago = 29 - i;
      return {
        d: i,
        v:
          ago < streak
            ? ago === 0
              ? 4
              : Math.min(4, ((i * 7 + 3) % 3) + 2)
            : ago < streak + 8 && (i * 13) % 4 === 0
              ? 1
              : 0,
      };
    }),
    timeline: (raw.timeline || []).map((b) => ({
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
    hasData: (raw.total_blocks || 0) > 0,
  };
}
