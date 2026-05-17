export type RawPayload = {
  period?: "today" | "week" | string;
  generated?: string;
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
  timeline?: {
    start: string;
    end: string;
    repo: string;
    branch: string;
    focus: number;
    switches: number;
    ai?: string;
  }[];
  streak_days?: number;
  top_repo?: string;
  top_branch?: string;
  ai_recap?: string;
  ai_per_block?: number;
  ai_tools?: ToolSummary[];
  machine?: string;
  os?: string;
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

declare global {
  interface Window {
    __FLOWD_DATA__?: RawPayload;
  }
}
