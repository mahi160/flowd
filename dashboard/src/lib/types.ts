export interface IFlowdData {
  date: string;
  period: "today" | "week" | "month";
}
// export type IFlowdData = {
//   period: string;
//   generated: string;
//   total_focus_min: number;
//   total_blocks: number;
//   total_switches: number;
//   files_changed: number;
//   lines_added: number;
//   lines_removed: number;
//   by_project: Record<string, number>;
//   by_tool: Record<string, number>;
//   languages: Record<string, number>;
//   heatmap: Heatmap[];
//   timeline: Timeline[];
//   streak_days: number;
//   top_repo: string;
//   top_branch: string;
//   ai_recap: string;
//   ai_per_block: number;
//   ai_sessions: AiSession[];
//   machine: string;
//   os: string;
// };
//
// interface Heatmap {
//   day: string;
//   hour: number;
//   minute: number;
// }
//
// interface Timeline {
//   start: string;
//   end: string;
//   repo: string;
//   branch: string;
//   focus: number;
//   switches: number;
// }
//
// interface AiSession {
//   tool: string;
//   project: string;
//   timestamp: string;
//   tokens_read: number;
//   tokens_write: number;
//   tokens_cache: number;
//   cost: number;
//   tools_called: number;
//   files_changed: number;
// }
