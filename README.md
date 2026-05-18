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

## 14:00 – 14:32

Focus: 30 min  ·  Switches: 4
Repo: flowd (main)
Projects: flowd 28min · scratch 2min
Commands: nvim 18min · claude 7min · git 5min
Languages: Go 16min · Markdown 2min
Code: 7 files (+142 −38)
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

1. **Git identity** — confirms your global `git config user.email` matches a verified address on GitHub. Required for commits to show on your contribution graph.
2. **Journal repo** — a private git repo where summaries are stored. Provide a remote URL (e.g. `git@github.com:you/journal.git`) or leave blank for local-only. If the remote exists, it will be cloned. If it doesn't exist yet, create an empty private repo on GitHub first, then run `fw init`.
3. **Machine name** — used as a subfolder in the journal repo (`macbook/`, `workstation/`, etc.). Defaults to your hostname. Lets you track multiple machines in one repo without conflicts.

At the end, `fw init` offers to add `fw start` to `~/.tmux.conf` so the daemon starts automatically when tmux starts. Recommended.

---

## Usage

```sh
fw start          # start daemon in background (or happens automatically with tmux)
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
- which command is running (`nvim`, `git`, `zsh`, etc.)
- branch name, files changed, lines added/removed

Once you've accumulated **30 focused minutes** (idle time doesn't count), it builds a block summary, appends it to your journal, and commits. Idle threshold defaults to 120 seconds — if your pane goes quiet for 2 minutes, that time isn't counted.

**Nothing is lost on restart.** The last block start time is persisted in the DB. If the daemon stops mid-session, it picks up from where it left off.

**Push schedule:** git push runs once per hour. Commits are timestamped when they're made, so GitHub shows your real work pattern regardless of when the push happens.

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

---

## Config

`~/.config/flowd/config.yaml`

```yaml
poll_interval_sec: 3      # how often to sample tmux
focus_block_min: 30       # focused minutes before writing a block
idle_threshold_sec: 120   # seconds of inactivity before pausing tracking
repo_path: ~/journal      # local path to your journal repo
git_remote: git@github.com:you/journal.git
branch: main
machine_name: macbook

# optional: AI summaries via any stdin→stdout CLI
ai_enabled: false
ai_command: "claude --print"
ai_prompt: "Summarize this coding session in 2 sentences."
```

### AI summaries

When `ai_enabled: true`, flowd pipes each block summary through `ai_command` and appends the output as a blockquote in the journal entry. Any tool that reads stdin and writes to stdout works:

```yaml
ai_command: "claude --print"        # Claude Code
ai_command: "pi -p --model haiku"   # pi
ai_command: "llm -m gpt-4o-mini"    # llm
```

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
