# Flowd

Local-first coding activity daemon for tmux workflows.

Flowd runs silently in the background, observes your tmux sessions, and builds a structured memory of how you spend your time — across projects, tools, and focus blocks. No cloud. No raw keylogging. Everything stored locally in SQLite.

```
fw init     # one-time setup
fw start    # start tracking
fw status   # check what's running
fw report today
```

---

## Requirements

- Go 1.22+ (to build from source)
- tmux (running when you work)
- git (optional, for log sync)
- macOS or Linux

---

## Install

**Build from source:**

```bash
git clone https://github.com/mahi/flowd
cd flowd
go build -o fw ./cmd/fw
sudo mv fw /usr/local/bin/fw   # or anywhere on your PATH
```

---

## Quick Start

### 1. Initialize

```bash
fw init
```

Walks you through setup interactively:

```
── Database ──────────────────────────────────────
  DB path [~/.local/share/flowd/flowd.db]:

── Polling ───────────────────────────────────────
  tmux poll interval (seconds) [3]:
  summary block interval (minutes) [30]:

── Keys ──────────────────────────────────────────
  track key counts (aggregated, no raw keys) (y/n) [y]:

── Git sync ──────────────────────────────────────
  sync logs to a private git repo (y/n) [n]:

── AI summaries ──────────────────────────────────
  Leave blank to skip. Examples: claude prompt / gemini chat
  AI command:

  config written → ~/.config/flowd/config.yaml
  database ready → ~/.local/share/flowd/flowd.db
  run `fw start` to begin tracking
```

### 2. Start the daemon

```bash
fw start
```

Runs in the foreground. Press `Ctrl+C` to stop.

> **Tip:** Add to your shell startup or tmux session via a background process:
> ```bash
> fw start &
> ```

### 3. Check status

```bash
fw status
```

```
db:     /Users/you/.local/share/flowd/flowd.db
events: 1420
blocks: 12
```

### 4. View reports

```bash
fw report today       # text table
fw report week        # last 7 days
fw report today --html > dashboard.html   # open in browser
```

### 5. Generate a summary on demand

```bash
fw summary now
```

```
## 14:00 – 14:30

**Repo:** flowd
**Focus:** 28 min
**Switches:** 6
**Tools:** editor, shell, git
```

---

## Commands

| Command | Description |
|---------|-------------|
| `fw init` | Interactive first-time setup |
| `fw init --force` | Re-run setup, overwrite existing config |
| `fw start` | Start the tracking daemon |
| `fw stop` | Stop the daemon *(coming soon)* |
| `fw status` | Show DB path, event count, block count |
| `fw summary now` | Build and print the current 30-min block |
| `fw report today` | Text report for today |
| `fw report week` | Text report for last 7 days |
| `fw report today --html` | HTML dashboard to stdout |
| `fw --debug start` | Verbose logging for troubleshooting |
| `fw --config <path> start` | Use a custom config file |

---

## Configuration

Config file: `~/.config/flowd/config.yaml`

```yaml
poll_interval_sec: 3          # how often tmux is queried
summary_interval_min: 30      # block generation cadence
track_keys: true              # aggregate key counts (no raw keys ever)
track_raw_keys: false         # must stay false
push_db: false                # enable git sync
repo_path: ~/flowd-private    # local repo for markdown logs
branch: main
ai_command: ""                # e.g. "claude prompt" or "gemini chat"
db_path: ~/.local/share/flowd/flowd.db
exclude_paths:
  - ~/.ssh
  - ~/Downloads
```

### AI summaries

Set `ai_command` to any CLI that reads a prompt from stdin and writes to stdout:

```yaml
ai_command: claude prompt
# or
ai_command: gemini chat
# or
ai_command: python ~/scripts/summarize.py
```

Flowd passes the block context as stdin. The output is stored as the block's summary.

### Git sync

Set `push_db: true` and point `repo_path` at a local git repo with a configured remote:

```bash
mkdir ~/flowd-private && cd ~/flowd-private
git init && git remote add origin git@github.com:you/flowd-private.git
```

After each 30-min block, Flowd writes a dated markdown file and pushes:

```
flowd-private/
  2026-04-27.md
  2026-04-28.md
  ...
```

---

## What Flowd tracks

| Signal | Stored as |
|--------|-----------|
| Active tmux pane | `pane_active` event every poll tick |
| Working directory change | `cwd_change` event |
| Command change | `cmd_change` event |
| Dominant project/repo | Inferred from `git rev-parse` on cwd |
| Tool category | Classified from command name |
| Focus time | Poll ticks × poll interval |
| Pane switches | Count of `cmd_change` events per block |

### Tool categories

| Command | Category |
|---------|----------|
| `nvim`, `vim` | `editor` |
| `lazygit`, `git` | `git` |
| `claude`, `gemini`, `aider`, `codex` | `ai` |
| `zsh`, `bash`, `fish` | `shell` |
| `node`, `python`, `go`, `cargo` | `runtime` |
| anything else | `other` |

---

## What Flowd does NOT do

- Store raw keystrokes
- Send data to any external service
- Track browser activity
- Capture screenshots
- Record audio or video
- Require a network connection

---

## Data location

| File | Path |
|------|------|
| Config | `~/.config/flowd/config.yaml` |
| Database | `~/.local/share/flowd/flowd.db` |
| Markdown logs | `~/flowd-private/YYYY-MM-DD.md` (if sync enabled) |

Inspect the database directly at any time:

```bash
sqlite3 ~/.local/share/flowd/flowd.db "SELECT * FROM events ORDER BY ts DESC LIMIT 20;"
```

---

## Troubleshooting

**tmux not detected**

Flowd logs a warning and idles until tmux is available. Make sure tmux is running before or after starting the daemon — it will pick up automatically on the next poll.

**No events recorded**

Run with `--debug` to see poll output:

```bash
fw --debug start
```

**Config not found**

Run `fw init` to create it. Or pass a custom path:

```bash
fw --config /path/to/config.yaml start
```

**Re-run init**

```bash
fw init --force
```

---

## Project structure

```
cmd/fw/             — CLI entry point
internal/
  config/           — config loader and writer
  db/               — SQLite open, WAL, migrations
  logger/           — structured logging (slog)
  collector/tmux/   — tmux pane state query
  collector/process/— git repo detection, command classification
  session/          — poll loop and event persistence
  summarizer/       — block aggregation and scheduler
  provider/         — AI command stdin adapter
  sync/             — markdown log writer, git push with retry
  report/           — text and HTML report generation
  initwizard/       — interactive setup wizard
```

---

## License

MIT
