# flowd

Tracks focused coding time in tmux and commits a summary to a private git repo every 30 minutes. Each commit lands on your GitHub contribution graph at the time you actually worked.

No cloud. No account. Data lives in SQLite and a git repo you own.

---

## The problem it solves

GitHub's contribution graph counts commits — not work. If you spend 4 hours debugging and push once at the end, that's one square. If you don't push at all, it's zero.

flowd fixes this by writing a structured journal entry every 30 focused minutes and committing it. The commit timestamp reflects when you worked, not when you pushed.

Sample journal entry:

```
### Monday, 12 May

**Total:** 60 min · 2 blocks · 3 context switches
**Top repo:** flowd · **Languages:** Go 45min · Shell 15min

> Worked on the proctree resolver and neovim plugin integration.
> Fixed language attribution for AI CLI panes.

## 14:00 – 14:32

**Focus:** 30 min  ·  **Context switches:** 2
**Repo:** flowd (main)
**Projects:** flowd 28min · scratch 2min
**Tools:** nvim 18min · pi 7min · git 5min
**Languages:** Go 16min · Markdown 2min
**Code:** 7 files (+142 −38)
```

---

## Requirements

- macOS or Linux
- tmux
- git (with a verified email on your GitHub account)

flowd only tracks tmux sessions. If you work outside tmux, it won't see it.

---

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/mahi160/flowd/main/install.sh | sh
```

This downloads a pre-built binary for your platform (Linux/macOS, amd64/arm64) and installs it to `/usr/local/bin/fw`.

---

## Setup

Run once after installing:

```sh
fw init
```

The wizard asks three things:

1. **Git identity** — confirms your global `git config user.email` matches a verified address on GitHub.
2. **Journal repo** — a private git repo where summaries are stored. Provide a remote URL or leave blank for local-only.
3. **Machine name** — subfolder in the journal repo (`macbook/`, `workstation/`, etc.). Defaults to your hostname.

At the end, `fw init` offers to add `fw start` to `~/.tmux.conf` so the daemon starts automatically when tmux starts. Recommended.

---

## Usage

```sh
fw start          # start daemon in background
fw stop           # stop the daemon
fw status         # daemon status + event/block counts
fw summary        # show the current in-progress block
fw summary --save # force-close the current block and write it now
fw report today   # text report for today
fw report week    # last 7 days
fw dashboard      # open HTML dashboard in browser
```

Logs: `~/.local/share/flowd/flowd.log`

---

## How it works

Every 3 seconds, flowd checks your active tmux pane and records:

- which git repo and project directory you're in
- which command is running (`nvim`, `git`, `pi`, `claude`, etc.)
- branch name, files changed, lines added/removed

**AI CLI detection:** if the pane shows `node` or `python` as the foreground
command, flowd walks the OS process tree to detect the real tool (e.g. `pi` or
`claude-code` running inside a Node interpreter). Time is attributed correctly
and no phantom JavaScript/Python "language" minutes appear.

Once you've accumulated **30 focused minutes**, it builds a block summary,
regenerates the day's journal section, and commits. Idle threshold defaults to
120 seconds.

**Nothing is lost on restart.** The last block start time is persisted in SQLite.

**Push schedule:** git push runs once per hour.

---

## Journal repo structure

```
your-journal/
  README.md              ← auto-updated monthly stats
  macbook/
    2026-04.md           ← monthly log
    2026-05.md
    flowd.db             ← SQLite (gitignored, stays local)
```

Each day's section has a roll-up (totals + AI standup) above the per-block entries.

---

## Config

`~/.config/flowd/config.yaml`  — see [docs/configuration.md](docs/configuration.md) for the full reference.

```yaml
poll_interval_sec: 3
focus_block_min:   30
idle_threshold_sec: 120
repo_path:   ~/journal
git_remote:  git@github.com:you/journal.git
branch:      main
machine_name: macbook

# AI: any stdin→stdout CLI works
ai_enabled: true
ai_command: "pi -p --model haiku"     # or: gemini -p --model gemini-2.0-flash
ai_prompt:  "Summarize this coding session in 2 short sentences."

# AI session transcript paths (for token/cost tracking)
ai_session_paths:
  claude-code: ~/.claude/projects
  pi:          ~/.pi/agent/sessions
```

### AI summaries & standup

When `ai_enabled: true`, flowd:

- Appends a per-block AI summary (blockquote) in the journal after each 30-min block.
- Generates a **today/yesterday standup** (git commits + tracked time per project) shown
  in the dashboard and as a roll-up header in the journal.

Any tool that reads stdin and writes to stdout works as `ai_command`:

```yaml
ai_command: "pi -p --model haiku"              # pi
ai_command: "claude --print"                   # Claude Code
ai_command: "gemini -p --model gemini-2.0-flash"  # Gemini
ai_command: "llm -m gpt-4o-mini"               # llm
```

---

## neovim plugin (optional)

The plugin is bundled in the binary. Install it with one command:

```sh
fw setup-nvim
```

Or choose "yes" when `fw init` asks (neovim is auto-detected).

The plugin writes the current filetype to `~/.local/share/flowd/nvim/<pid>.json`
on every buffer switch. The daemon reads it for nvim panes and falls back
gracefully when the plugin is absent. See [docs/nvim-plugin.md](docs/nvim-plugin.md).

---

## Docs

- [Architecture](docs/architecture.md)
- [How it works](docs/how-it-works.md)
- [Data model](docs/data-model.md)
- [neovim plugin](docs/nvim-plugin.md)
- [Configuration](docs/configuration.md)
- [ADR: process-tree AI detection](docs/adr/0001-process-tree-ai-detection.md)
- [ADR: nvim plugin file contract](docs/adr/0002-nvim-plugin-file-contract.md)
- [ADR: static dashboard embed](docs/adr/0003-static-dashboard-embed.md)

---

## Uninstall

```sh
fw stop
rm /usr/local/bin/fw
rm -rf ~/.config/flowd
rm -rf ~/.local/share/flowd
# optionally delete your journal repo
# remove the fw start line from ~/.tmux.conf
```
