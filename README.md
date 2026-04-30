# flowd

Local coding-activity tracker. Watches your tmux sessions, builds focus-based
summaries, and commits them to a private git journal — every commit is a green
square on your GitHub profile.

No cloud. No telemetry. Your data lives in a SQLite file and a private git repo
you own.

---

## What it does

Every few seconds it checks your **attached** tmux session and records:

- which project (cwd / git repo)
- which command / tool (`nvim`, `claude`, `zsh`, `git`, …)
- branch, files changed, lines +/−
- context switches

Once you've accumulated **30 focused minutes** (idle time doesn't count), it
closes a block, writes a summary into your journal repo, and commits it.
Each commit shows on your GitHub contribution graph at the time you actually
worked.

Sample journal entry:

```md
### Monday, 30 Apr

## 14:00 – 16:42

**Focus:** 30 min  ·  **Switches:** 4
**Repo:** flowd (main)
**Projects:** flowd 28min · scratch 2min
**Commands:** nvim 18min · claude 7min · git 5min
**Languages:** Go 16min · Markdown 2min
**Code:** 7 files (+142 −38)
```

---

## Install

```sh
git clone https://github.com/mahi160/flowd
cd flowd
go build -o fw ./cmd/fw
sudo mv fw /usr/local/bin/
```

Requires **Go 1.22+**, **tmux**, **git**.

---

## Quickstart

```sh
fw init          # one-time setup (2 minutes)
fw start         # run the daemon (foreground)
fw status        # check DB counts
fw summary       # preview the current in-progress block
fw report today  # text report for today
fw report week   # last 7 days
fw dashboard     # open HTML dashboard in browser
```

`fw init` offers to add `fw start` to `~/.tmux.conf` so the daemon
starts automatically with tmux.

---

## Journal repo structure

```
flowd-private/              ← your private git repo
  README.md                 ← auto-generated monthly stats
  macbook/                  ← one folder per machine
    2026-04.md              ← monthly markdown journal
    2026-05.md
    flowd.db                ← SQLite (backed up via git)
```

Each machine writes to its own subfolder — no conflicts when you add a
second machine later.

---

## How it works

```
  every 3s ──▶  tmux list-clients
                    │ attached?
                    ▼
               active pane → command, cwd, repo
               classifyCommand(cmd) → category (for lang inference)
               insert pane_active event in SQLite

  every 30s ──▶  count focused minutes since last block
                    │ ≥ 30?
                    ▼
               BuildBlock → aggregate commands, projects, langs, git diff
               WriteJournal → <machine>/YYYY-MM.md
               WriteReadme  → README.md (monthly stats)
               git commit
               (push only on startup + 10 pm)
```

**Nothing is lost on restart.** `blockStart` is persisted in the DB's
`state` table. The daemon picks up exactly where it left off, even after
a reboot.

**Push schedule:** git push runs twice — once when the daemon starts (to
sync any commits from previous sessions) and once at 10 pm local time.
Commits are timestamped when they're made, so GitHub shows your actual
work pattern regardless of when the push happens.

---

## Config

`~/.config/flowd/config.yaml`

```yaml
poll_interval_sec: 3        # how often to sample tmux
focus_block_min: 30         # focused minutes per block
idle_threshold_sec: 120     # pause tracking after N sec no input
repo_path: ~/flowd-private  # journal repo root
git_remote: git@github.com:you/flowd-private.git
branch: main
machine_name: macbook       # subfolder name (default: hostname)
watch_dirs:
  - ~

# AI summaries (optional — any stdin→stdout CLI)
ai_enabled: true
ai_command: "pi -p --model haiku"
ai_prompt: "Summarize this coding session (30 focused minutes) in 2 short sentences."
```

### AI integration

Flowd pipes each block summary through any CLI tool that reads stdin and
prints to stdout:

| Tool | `ai_command` |
|---|---|
| pi (haiku) | `pi -p --model haiku` |
| Claude Code | `claude --print` |
| llm | `llm -m claude-3-5-sonnet` |
| aider | `aider --msg -` |

---

## Commands

| Command | What it does |
|---|---|
| `fw init` | Interactive setup |
| `fw start` | Run daemon in foreground |
| `fw stop` | Stop the running daemon |
| `fw status` | Daemon status + DB counts |
| `fw summary` | Preview current in-progress block |
| `fw summary --save` | Force-close the current block |
| `fw report today` | Text report for today |
| `fw report week` | Last 7 days |
| `fw dashboard [today\|week]` | HTML dashboard in browser |
| `fw setup-tmux` | Add `fw start` to `~/.tmux.conf` |

---

## GitHub contribution graph

flowd is designed around the GitHub contribution graph. Every 30 focused
minutes = one commit = one green square. The commit timestamp reflects when
you actually worked, not when it was pushed.

**Important:** the email in your global git config must match a verified
address on your GitHub account. `fw init` checks this for you.

---

## Privacy

- All data is local. Only the journal repo is pushed, to the remote you set.
- File **contents** are never read — only filenames, extensions, line counts.
- Polling stops when tmux has no attached client.
- The DB is inside your private journal repo and backed up via git automatically.

---

## Uninstall

```sh
rm /usr/local/bin/fw
rm -rf ~/.config/flowd
# journal repo at ~/flowd-private — delete if you want
# remove the run-shell line from ~/.tmux.conf
```
