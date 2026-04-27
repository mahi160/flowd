# Product Requirements Document — Flowd

## Overview

Flowd is a local-first developer activity daemon for tmux-based workflows. It observes your terminal sessions, stores structured telemetry in SQLite, generates 30-minute work summaries, and optionally syncs logs to a private Git repository.

No cloud. No surveillance. No raw keystroke capture. Your data stays on your machine.

---

## Problem Statement

Solo developers working in tmux have no passive memory of how their time is spent. Switching between Neovim, shell, lazygit, and AI CLIs produces no durable record of focus, context, or progress. At end of day, it is impossible to reconstruct what was worked on, for how long, or in what order.

### Specific pain points

- No record of how coding time was distributed across projects
- No visibility into focus vs. distraction patterns
- Lost work context between sessions and across days
- No lightweight work journal without manual effort
- No insight into which tools dominate workflow time

---

## Goals

| Goal | Metric |
|------|--------|
| Passive activity tracking | Automatic, zero-friction |
| Low resource overhead | CPU < 1%, RAM < 50 MB |
| Private by default | All data local, no external calls |
| Useful daily summaries | Accurate 30-min blocks |
| Repo awareness | Correct project/repo detection |

---

## Non-Goals

- Employee surveillance or monitoring
- Raw keystroke logging
- Cloud SaaS or telemetry upload
- Screenshot or screen recording
- Browser activity tracking
- Gamified productivity scores
- Multi-user support

---

## User

Solo developer using:

- macOS or Linux
- tmux as primary terminal multiplexer
- Neovim as primary editor
- lazygit for version control
- Shell (zsh/bash/fish) for commands
- AI CLI tools (Claude, Gemini, Codex, Aider)

---

## MVP Feature Set

### 1. tmux Activity Tracking
- Poll active pane every N seconds (default 3)
- Capture: session name, window name, pane index, current command, working directory
- Classify commands into categories: `editor`, `git`, `ai`, `shell`, `runtime`, `other`
- Detect cwd changes and command switches as discrete events

### 2. Event Storage
- All events written to local SQLite database
- WAL mode for safe concurrent reads
- Schema: `events` table (id, ts, type, value, meta JSON)

### 3. 30-Minute Block Generation
- Aggregate events into aligned 30-min blocks
- Per block: focused minutes, pane switches, tools used, dominant repo/project
- Schema: `blocks` table with all aggregated fields
- Scheduler fires on clock-aligned boundaries (e.g. :00, :30)

### 4. Summaries
- Template-based markdown summary generated for each block (no AI required)
- Optional: pass summary context to a user-configured AI CLI via stdin
- AI command is fully user-controlled (any CLI that reads stdin)

### 5. Git Sync
- Write dated markdown logs to a local private repo (`2026-04-27.md`)
- Commit and push after each block (optional, `push_db: true`)
- Exponential backoff retry on push failure

### 6. Reports
- `fw report today` — text table of all blocks for today
- `fw report week` — last 7 days
- `fw report today --html` — dark-mode HTML dashboard written to stdout

---

## Data Model

```sql
CREATE TABLE events (
  id    INTEGER PRIMARY KEY AUTOINCREMENT,
  ts    DATETIME NOT NULL,
  type  TEXT NOT NULL,     -- pane_active | cwd_change | cmd_change
  value TEXT NOT NULL,
  meta  TEXT NOT NULL      -- JSON: session, window, pane, command, category, cwd, repo
);

CREATE TABLE blocks (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  start_ts        DATETIME NOT NULL,
  end_ts          DATETIME NOT NULL,
  project         TEXT NOT NULL,
  repo            TEXT NOT NULL,
  focused_minutes INTEGER NOT NULL,
  key_count       INTEGER NOT NULL,
  switches        INTEGER NOT NULL,
  tools           TEXT NOT NULL,   -- JSON array
  summary         TEXT NOT NULL    -- markdown
);
```

---

## Configuration

All settings in `~/.config/flowd/config.yaml`.

| Key | Default | Description |
|-----|---------|-------------|
| `poll_interval_sec` | `3` | tmux poll frequency |
| `summary_interval_min` | `30` | block generation interval |
| `track_keys` | `true` | aggregate key count buckets |
| `track_raw_keys` | `false` | must stay false — raw keys never stored |
| `push_db` | `false` | enable git sync |
| `repo_path` | `~/flowd-private` | local git repo for logs |
| `branch` | `main` | sync branch |
| `ai_command` | `` | external AI CLI (reads from stdin) |
| `db_path` | `~/.local/share/flowd/flowd.db` | SQLite file |
| `exclude_paths` | `[]` | paths to suppress from tracking |

---

## Privacy Constraints

- `track_raw_keys` is present in config but the collector never logs raw keystrokes regardless of value
- AI calls are made only via the user-configured `ai_command`; no data is sent to any hardcoded endpoint
- Git sync pushes only to a repo the user configures — no default remote
- All data is readable by the user in plain SQLite and plain markdown

---

## Architecture

```
cmd/fw
internal/
  config/        — yaml loader, defaults, writer
  db/            — sqlite open, WAL, schema migrations
  logger/        — slog wrapper
  collector/
    tmux/        — active pane query via tmux display-message
    process/     — git repo detection, command classification
  session/       — poll loop, event writer
  summarizer/    — block aggregation, scheduler, template summaries
  provider/      — stdin-based AI command adapter
  sync/          — markdown log writer, git commit+push with retry
  report/        — text and HTML report generators
  initwizard/    — interactive setup wizard
```

---

## Future Scope (Post-MVP)

- `fw stop` — PID file daemon management
- Key count collection via tmux hooks (not polling)
- `fw report week --html` — weekly HTML dashboard
- Neovim plugin integration for file-level tracking
- Project tagging and manual annotations
- `fw log` — tail live event stream
