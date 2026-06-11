# How it works

## Tracking loop

Every `poll_interval_sec` (default 3 s), the Tracker:

1. Calls `tmux list-clients` to find the most-recently-active attached client
   and its idle time. If the client is idle beyond `idle_threshold_sec`
   (default 120 s), tracking pauses.
2. Calls `tmux display-message` on the active session to get the pane's
   current working directory, command name, and PID.
3. If the command is a **known interpreter** (`node`, `python`, `bun`, …),
   walks the process tree from the pane PID to detect AI CLIs running inside
   (e.g. `pi` or `claude` wrapped in a Node runtime). See
   [ADR-0001](adr/0001-process-tree-ai-detection.md).
4. If the command is `nvim` and the
   [flowd.lua plugin](nvim-plugin.md) is installed, reads the plugin's state
   file for the active filetype — gives more accurate language attribution
   before a commit lands. See [ADR-0002](adr/0002-nvim-plugin-file-contract.md).
5. Writes a `pane_active` event to SQLite with a JSON payload
   (`session`, `cwd`, `command`, `category`, `repo`, `nvim_ft`, …).
6. Writes `cwd_change` / `session_change` events when the directory or
   session switches.

## Focus block cadence

A separate scheduler goroutine ticks every minute, counting `pane_active`
events since the last block boundary. When the count × poll interval ≥
`focus_block_min` minutes (default 30):

1. **BuildBlock** aggregates all events in the window:
   - `ByTool` — seconds per command (nvim, git, zsh, pi, …) → minutes
   - `ByProject` — seconds per repo → minutes
   - `Languages` — from git diff stats (lines weighted), nvim plugin
     filetype, or cwd file scan, in that priority order
   - `FilesAdded`, `LinesAdded`, `LinesDel` — from `git log --numstat`
2. (Optional) `RunAI` pipes the block summary through `ai_command` for a
   per-block AI summary.
3. **WriteJournal** regenerates the current day's section in the monthly
   markdown file — a roll-up header (total min + AI standup) above the
   per-block entries.
4. **CommitJournal** runs `git add . && git commit` in the journal repo.
   The commit message is `flowd: <ISO-timestamp>`.

## Push schedule

`git push` runs once per hour (with `--rebase` pull before push and
3 retries with exponential backoff).

## AI sessions (pi / claude-code)

Separately from tmux tracking, a goroutine scans the JSONL transcript
directories (e.g. `~/.pi/agent/sessions`, `~/.claude/projects`) every 30
minutes. Each file is processed by a registered `Processor` (one per tool)
and results land in the `ai_sessions_raw` table with token counts, cost, and
session metadata. This is the authoritative source for AI usage stats — the
tmux tracking only determines *time attribution*.

## Dashboard

`fw dashboard` is a read-only command that:

1. Loads blocks for six periods (today, yesterday, week, month, year, all)
   from SQLite.
2. Builds a standup (today + yesterday activity + git commits per project,
   fed through `ai_command`; result is cached by a hash of the input blocks).
3. Serialises everything to a JSON payload and injects it into the embedded
   Svelte dashboard HTML.
4. Writes the file to `$TMPDIR/flowd.html` and opens it in the default browser.

The dashboard is a static snapshot — refreshing the page won't fetch new data.
Re-run `fw dashboard` to get updated numbers.
