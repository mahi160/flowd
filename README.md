# Flowd

Local coding-activity tracker. Watches your tmux sessions, records what you
worked on, and writes a private markdown journal — like Wakatime, but
self-hosted and powered by tmux.

No cloud. No telemetry. Your data lives in a SQLite file and (optionally) a
private git repo you own.

---

## What it does

Every few seconds it checks your **attached** tmux session and records:

- which project (cwd / git repo)
- which tool (editor, ai, git, shell, runtime)
- branch, files changed, lines +/−
- pane/cmd switches

Every 30 minutes it builds a block, writes a Wakatime-style summary into
`YYYY-MM.md` in your journal repo, and (if enabled) commits + pushes.

Sample journal entry:

```md
## 14:00 – 14:30

**Focus:** 22 min · **Switches:** 4
**Repo:** flowd (main)
**Projects:** flowd 18min · scratch 4min
**Tools:** editor 15min · ai 4min · git 3min
**Languages:** Go 16min · Markdown 2min
**Code:** 7 files (+142 −38)
```

---

## Install

```sh
git clone <this repo>
cd flowd
go build -o fw ./cmd/fw
sudo mv fw /usr/local/bin/        # or anywhere on $PATH
```

Requires Go 1.26+, tmux, git, sqlite (bundled via go-sqlite3).

---

## Quickstart

```sh
fw init          # interactive setup
fw start         # run daemon (foreground)
fw status        # check db
fw summary       # build + print last block
fw report today  # text report for today
fw report week   # last 7 days
```

`fw init` will offer to add `fw start` to your `~/.tmux.conf` so the daemon
starts automatically with tmux.

---

## How it works

```
                         ┌───────────────────────┐
  every 3s ──────────▶   │   tmux list-clients   │
                         └──────────┬────────────┘
                                    │
                       attached?    │  no  ──▶ idle, skip
                                    │ yes
                                    ▼
                       ┌──────────────────────────┐
                       │ tmux display-message -t  │  → active pane
                       │ session, window, cmd,    │
                       │ cwd                      │
                       └──────────┬───────────────┘
                                  ▼
                         classify cmd → tool category
                         git → repo, branch
                         insert event row in SQLite

  every 30 min ─▶  scan window's events
                   ├─ count active ticks → focused minutes
                   ├─ aggregate per tool, per project
                   ├─ git log + git diff → files, lines, langs
                   └─ render markdown
                          │
                   focused ≥ min_focus_min ?
                          │ yes
                          ▼
                   append to journal/YYYY-MM.md
                   git commit + push (if enabled)
```

### Algorithm details

**Polling (every `poll_interval_sec`, default 3s):**

1. Run `tmux list-clients`. If empty → user is detached, **do nothing**.
2. Run `tmux display-message -t <attached_session>` → grab session, window,
   pane id, command, cwd.
3. Skip if cwd is not under any `watch_dirs`.
4. Resolve git repo name + classify command → `editor` / `ai` / `git` /
   `shell` / `runtime` / `other`.
5. Insert one `pane_active` event row. Add `cwd_change` / `cmd_change` rows
   if changed since last tick.

**Block builder (every `summary_interval_min`, default 30m, on clock
boundary):**

1. Read every event in `[start, end)`.
2. `focused_min = active_ticks * poll_interval_sec / 60`.
3. Per-tool / per-project minutes = sum of ticks tagged with that
   category / repo, converted to minutes.
4. For each repo touched:
   - `git log --since --until --numstat` → committed lines/files in window.
   - `git diff --numstat` → uncommitted lines/files.
   - File extensions of changed files → languages, weighted by editor time
     in that repo.
5. Render markdown summary.
6. **Push gate:** if `focused_min < min_focus_min` (default 15), save the
   block to the DB but **skip** journal write + git push. Stops empty/idle
   half-hours from polluting your journal.

---

## Config

`~/.config/flowd/config.yaml`

```yaml
poll_interval_sec: 3 # how often to sample tmux
summary_interval_min: 30 # block size
min_focus_min: 15 # below this → no journal/push
push_db: true # commit + push journal repo
repo_path: ~/flowd-private # local journal repo
git_remote: git@github.com:you/my-journal.git
branch: main
db_path: ~/.local/share/flowd/flowd.db
watch_dirs:
  - ~/code
  - ~/work
```

`fw init` writes this for you. Edit by hand any time.

### What's tracked vs. ignored

Tracked: tmux session/window/pane, command name, cwd, git repo + branch,
file diffs (numstat).

**Not** tracked: keystrokes, file _contents_, anything outside `watch_dirs`,
anything when no tmux client is attached.

---

## Journal repo

A separate, **private** git repo (NOT your project repo) where flowd writes
summaries. One file per month: `2026-04.md`, `2026-05.md`, …

Each block appends under a day heading. Pushed on the same 30-min cadence
when the block has real activity.

To set up:

```sh
# during fw init, answer "yes" to journal repo and paste a remote URL.
# OR manually:
mkdir ~/flowd-private && cd ~/flowd-private
git init -b main
git remote add origin git@github.com:you/my-journal.git
```

---

## Commands

| Command           | What it does                                 |
| ----------------- | -------------------------------------------- |
| `fw init`         | Interactive config setup                     |
| `fw start`        | Run daemon in foreground                     |
| `fw status`       | Print DB path + event/block counts           |
| `fw summary`      | Build the most recent 30-min block, print it |
| `fw report today` | Text report for today                        |
| `fw report week`  | Text report for last 7 days                  |
| `fw setup-tmux`   | Add `fw start` to `~/.tmux.conf`             |

Flags: `--config <path>`, `--debug`.

---

## Privacy

- All data is local. Only the journal repo is pushed, and only to the
  remote you set.
- File contents are never read. Only filenames, extensions, line counts.
- Polling stops when tmux has no attached client.

---

## Uninstall

```sh
rm /usr/local/bin/fw
rm -rf ~/.config/flowd ~/.local/share/flowd
# remove the run-shell line from ~/.tmux.conf
# (journal repo at ~/flowd-private stays — delete if you want)
```
