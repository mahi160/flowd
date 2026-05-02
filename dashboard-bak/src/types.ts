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
  timeline?: { start: string; end: string; repo: string; branch: string; focus: number; switches: number; ai?: string }[];
  streak_days?: number;
  top_repo?: string;
  top_branch?: string;
  ai_recap?: string;
  ai_per_block?: number;
  ai_sessions?: {
    tool: string;
    project: string;
    timestamp: string;
    tokens_read: number;
    tokens_write: number;
    tokens_cache: number;
    cost: number;
    tools_called: number;
    files_changed: number;
  }[];
  machine?: string;
  os?: string;
};

export type Item = { name: string; min: number; color: string };
export type Data = ReturnType<typeof import("./lib/data").transform>;

declare global {
  interface Window {
    __FLOWD_DATA__?: RawPayload;
  }
}
