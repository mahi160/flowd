export type IHourBucket = {
  day: string; // "Mon 22" or "Today"
  hour: number; // 0–23 (week) or 0–47 half-hour index (today)
  minute: number; // focused minutes in the bucket
};

export type ITlBlock = {
  start: string;
  end: string;
  repo: string;
  branch: string;
  focus: number;
  switches: number;
  ai?: string;
};

export type IRawPayload = {
  period: "today" | "week" | string;
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
  heatmap: IHourBucket[];
  timeline: ITlBlock[];
  streak_days: number;
  top_repo: string;
  top_branch: string;
  ai_recap: string;
  ai_per_block: number;
  machine: string;
  os: string;
};
