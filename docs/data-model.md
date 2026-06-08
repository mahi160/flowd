# Data model

All persistent data lives in a SQLite file at
`<repo_path>/<machine_name>/flowd.db` (gitignored — stays local).

## Tables

### `events`

The raw, append-only signal from the tracker.

| Column | Type    | Notes |
|--------|---------|-------|
| `id`   | INTEGER | PK, autoincrement |
| `ts`   | TEXT    | ISO 8601 UTC timestamp |
| `type` | TEXT    | `pane_active` \| `cwd_change` \| `session_change` |
| `value`| TEXT    | pane ID (active), new cwd (cwd_change), session name (session_change) |
| `meta` | TEXT    | JSON `PaneMeta`: session, window, pane, command, category, cwd, repo, machine, os, nvim_ft |

Events older than 90 days are pruned by the daemon's daily cleanup task.
Blocks are built from events, so the raw events are expendable after
block-building — the 90-day window gives a generous buffer.

### `blocks`

One row per completed focus block (30 focused minutes of work).

| Column           | Type    | Notes |
|------------------|---------|-------|
| `id`             | INTEGER | PK |
| `start_ts`       | TEXT    | ISO 8601 UTC — also a UNIQUE constraint |
| `end_ts`         | TEXT    | ISO 8601 UTC |
| `project`        | TEXT    | Dominant cwd basename |
| `repo`           | TEXT    | Dominant git repo name |
| `focused_minutes`| INTEGER | Wall-clock focused minutes (idle excluded) |
| `switches`       | INTEGER | Context switches within the block |
| `data`           | TEXT    | Full `Block` struct as JSON |
| `summary`        | TEXT    | Pre-rendered markdown summary text |
| `ai_summary`     | TEXT    | Optional per-block AI summary (never overwritten once written) |

### `ai_sessions_raw`

One row per message in an AI tool's JSONL transcript.

| Column         | Type    | Notes |
|----------------|---------|-------|
| `id`           | INTEGER | PK |
| `tool`         | TEXT    | `pi` \| `claude-code` \| … |
| `project`      | TEXT    | From the session's cwd |
| `session_id`   | TEXT    | Tool-specific session UUID |
| `model`        | TEXT    | Model string from the transcript |
| `ts`           | TEXT    | Message timestamp |
| `tokens_read`  | INTEGER | Input + cache-read tokens |
| `tokens_write` | INTEGER | Output tokens |
| `tokens_cache` | INTEGER | Cache-creation tokens |
| `cost`         | REAL    | Cost in USD (0 when not provided by tool) |

### `ai_sessions_watermark`

One row per JSONL file, tracking how many bytes have been ingested so
incremental scans only process new content.

| Column       | Type    | Notes |
|--------------|---------|-------|
| `path`       | TEXT    | Absolute file path (PK) |
| `offset`     | INTEGER | Bytes processed so far |
| `updated_at` | TEXT    | Last scan timestamp |

### `state`

Key/value store for daemon persistence across restarts.

| Column | Type | Notes |
|--------|------|-------|
| `key`  | TEXT | PK |
| `val`  | TEXT | |

Current keys:
- `block_start_ts` — the start time of the current in-progress block;
  persisted so a daemon restart picks up from where it left off.

## Journal repo

```
<repo_path>/
  README.md              ← auto-regenerated monthly stats
  <machine_name>/
    flowd.db             ← SQLite (gitignored)
    2026-04.md           ← monthly log
    2026-05.md
```

Each monthly `.md` has this structure:

```markdown
# May 2026

### Monday, 05 May

**Total:** 90 min · 3 blocks · 5 context switches
**Top repo:** flowd · **Languages:** Go 70min · Shell 20min

> Today I worked on the proctree resolver and nvim plugin. Fixed a bug in
> the block builder's language inference.

## 14:00 – 14:32

**Focus:** 30 min  ·  **Context switches:** 2
**Repo:** flowd (main)
...
```
