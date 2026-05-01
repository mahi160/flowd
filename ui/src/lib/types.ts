// ── Raw payload from Go (snake_case JSON) ────────────────────────────────────

export type HourBucket = {
  day: string;    // "Mon 22" or "Today"
  hour: number;   // 0–23 (week) or 0–47 half-hour index (today)
  minute: number; // focused minutes in the bucket
};

export type TlBlock = {
  start: string;
  end: string;
  repo: string;
  branch: string;
  focus: number;
  switches: number;
  ai?: string;
};

export type RawPayload = {
  period: 'today' | 'week' | string;
  generated: string;
  total_focus_min: number;
  total_blocks: number;
  total_switches: number;
  files_changed: number;
  lines_added: number;
  lines_removed: number;
  by_project: Record<string, number>;
  by_tool: Record<string, number>;
  languages: Record<string, number>;
  heatmap: HourBucket[];
  timeline: TlBlock[];
  streak_days: number;
  top_repo: string;
  top_branch: string;
  ai_recap: string;
  ai_per_block: number;
  machine: string;
  os: string;
};

// ── Transformed / parsed types ────────────────────────────────────────────────

export type Item = { name: string; min: number; color: string };

export type TimelineEntry = {
  from: string;
  to: string;
  project: string | null;
  branch: string | null;
  focus: number;
  switches: number;
  ai: string | null;
};

export type WeekDay = { day: string; date: string; min: number };
export type StreakCell = { d: number; v: number };

export type ParsedData = {
  period: 'today' | 'week';
  generated: string;
  machine: string;
  os: string;
  topRepo: { name: string; branch: string };
  focus: { totalMin: number; blocks: number; switches: number };
  code: { files: number; added: number; removed: number };
  byProject: Item[];
  byCommand: Item[];
  byLanguage: Item[];
  hourly: number[];                              // 24 values (hour 0–23)
  weekHourlyByDay: Record<string, number[]>;     // day label → 24 values
  weekDays: WeekDay[];
  streakDays: number;
  streakCells: StreakCell[];
  timeline: TimelineEntry[];
  aiRecap: string | null;
  aiPerBlock: number;
  hasData: boolean;
};
