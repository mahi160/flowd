export type CalDay = {
  date: string;   // "2026-05-17"
  dow: number;    // 0=Sun … 6=Sat (Go time.Weekday)
  min: number;
  blocks: number;
};

export type MonthBar = {
  ym: string;    // "2026-05"
  year: number;
  month: number; // 1-12
  min: number;
  blocks: number;
};

export type ToolSummary = {
  tool: string;
  total_cost: number;
  total_input: number;
  total_output: number;
  total_cache: number;
  session_count: number;
  message_count: number;
  top_model: string;
  model_breakdown: Record<string, number>;
  sessions: AggregatedSession[];
};

export type AggregatedSession = {
  tool: string;
  project: string;
  session_id: string;
  model: string;
  start_time: string;
  end_time: string;
  start_unix: number;
  total_input: number;
  total_output: number;
  total_cache: number;
  total_cost: number;
  message_count: number;
};

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

export type DayBucket = { day: string; date: string; min: number };

// ── Per-period data ─────────────────────────────────────────────────────────
export type RawPeriodData = {
  total_focus_min?: number;
  total_blocks?: number;
  total_switches?: number;
  files_changed?: number;
  lines_added?: number;
  lines_removed?: number;
  by_project?: Record<string, number>;
  by_tool?: Record<string, number>;
  languages?: Record<string, number>;
  heatmap?: { day: string; hour: number; minute: number }[];
  cal_days?: CalDay[];
  month_bars?: MonthBar[];
  timeline?: {
    start: string; end: string; repo: string; branch: string;
    focus: number; switches: number; ai?: string;
  }[];
  top_repo?: string;
  top_branch?: string;
  ai_tools?: ToolSummary[];
  tracking_since?: string;
  active_days?: number;
  best_day_date?: string;
  best_day_min?: number;
};

// ── Top-level payload ───────────────────────────────────────────────────────
export type RawPayload = {
  initial_period?: string;
  generated?: string;
  machine?: string;
  os?: string;
  streak_days?: number;
  /** AI-generated today/yesterday standup text. */
  standup?: string;
  /** Structured standup input (always present when there is recent activity). */
  standup_raw?: string;
  periods?: Record<string, RawPeriodData>;
};

declare global {
  interface Window {
    __FLOWD_DATA__?: RawPayload;
  }
}
