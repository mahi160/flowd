# Configuration

Config file: `~/.config/flowd/config.yaml`

Run `fw init` to generate it interactively, or create it manually.

## Full reference

```yaml
# ── Tracking ──────────────────────────────────────────────────────────────
poll_interval_sec: 3       # how often to sample the active tmux pane
focus_block_min:   30      # focused minutes that trigger one block + commit
idle_threshold_sec: 120    # pane-quiet seconds before pausing tracking

# Directories to track. Only panes whose cwd is inside a watch_dir are
# recorded. Defaults to your home directory if omitted.
watch_dirs:
  - ~/code
  - ~/work

# ── Journal repo ──────────────────────────────────────────────────────────
repo_path:    ~/journal    # local path to the private journal git repo
git_remote:   git@github.com:you/journal.git  # blank = local-only (no push)
branch:       main
machine_name: macbook      # subfolder inside repo (default: hostname)

# ── AI ────────────────────────────────────────────────────────────────────
# Any CLI that reads stdin and writes to stdout works.
ai_enabled: true
ai_command: "pi -p --model haiku"   # or: claude --print / llm -m gpt-4o-mini
ai_prompt: >
  Summarize this coding session (30 focused minutes) in 2 short sentences.
  Focus on what was accomplished. Be concise.

# AI session transcript paths (for token/cost tracking).
# These are the directories each tool writes its JSONL session files to.
ai_session_paths:
  claude-code: ~/.claude/projects
  pi:          ~/.pi/agent/sessions
```

## Standup AI command

The standup (today/yesterday summary shown in the dashboard and journal)
uses the same `ai_command` as per-block summaries. To use a faster/cheaper
model just for standup, you can change `ai_command` — both features share it.

Recommended standup commands:

| Tool | Command |
|------|---------|
| pi | `pi -p --model haiku` |
| claude-code | `claude --print --model claude-haiku-4-5` |
| Gemini | `gemini -p --model gemini-2.0-flash` |
| llm | `llm -m gpt-4o-mini` |

## Idle detection

On **macOS**, flowd also pauses tracking when the laptop lid is closed
(detected via IOKit's `AppleClamshellState`).

On **Linux**, only `idle_threshold_sec` is used. If your screen locks and
you step away, any gap longer than `idle_threshold_sec` is automatically
excluded from focused-minute counting.

## Adding a new AI tool

To track a new AI tool's token usage (not just tmux time), implement the
`ai_sessions.Processor` interface and register it with `Register()`.
See `internal/ai_sessions/pi.go` for a reference implementation.
