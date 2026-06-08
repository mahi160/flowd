# Architecture

flowd is a single Go binary (`fw`) with three long-running goroutines and a
Svelte 5 dashboard compiled to a single embedded HTML file.

```
fw (binary)
├── cmd/fw/main.go          — CLI entry point (cobra)
└── internal/
    ├── fw/                 — core daemon
    │   ├── daemon.go       — process lifecycle, signal handling, goroutine wiring
    │   ├── track.go        — tmux poller (Tracker) + proctree AI-CLI resolver
    │   ├── proctree.go     — walks /proc (Linux) or ps (macOS) to detect AI CLIs
    │   ├── nvim.go         — reads plugin/flowd.lua state files for nvim panes
    │   ├── block.go        — event aggregation → Block (30-min focus summary)
    │   ├── journal.go      — regenerates monthly markdown + README
    │   ├── standup.go      — builds AI standup from git commits + block data
    │   ├── git.go          — git helpers: DiffStat, RecentCommits, ScanLangs, …
    │   ├── dashboard.go    — builds JSON payload, renders embedded HTML
    │   ├── commands.go     — cobra commands: dashboard, summary, report, …
    │   ├── scheduler.go    — focus/push/cleanup cadences (inside daemon.go)
    │   ├── db.go           — SQLite schema + helpers
    │   ├── config.go       — YAML config load/save
    │   └── static/
    │       └── dashboard.html  ← compiled Svelte app (go:embed)
    └── ai_sessions/        — pluggable AI-tool JSONL processors
        ├── ai_sessions.go  — Processor interface + registry
        ├── service.go      — scans directories, writes ai_sessions_raw
        ├── pi.go           — pi processor
        └── claude.go       — claude-code processor

dashboard/                  — Svelte 5 + Tailwind 4 + Vite source
plugin/
└── flowd.lua               — optional neovim plugin
```

## Three goroutines inside the daemon

```
runDaemon()
  ├─ go Tracker.Run(ctx)          — polls tmux every 3 s, writes events
  ├─ go runScheduler(ctx, …)      — drives block/commit/push cadences
  └─ go ai_sessions.Service loop  — scans JSONL dirs every 30 min
```

## Data flow

```
tmux pane
  → Tracker.poll()
      → proctree.ResolveAICLI()   (only when interpreter detected)
      → NvimFiletype()            (only when nvim pane + plugin present)
      → INSERT INTO events
            (ts, type=pane_active, value=pane_id, meta=PaneMeta JSON)

every minute → countFocusedMin()
  → if >= focusBlockMin:
      BuildBlock()               — aggregates events → Block
        → DiffStat()             — git diff for language/code stats
        → distributeByLines()    — language attribution from changed files
        → nvimLangTicks          — plugin filetype overrides git-diff
      RunAI()                    — optional per-block AI summary (stdin→stdout)
      WriteJournal()             — regenerates day section in monthly .md
        → BuildStandup()         — today/yesterday AI standup (cached)
      CommitJournal()            — git add + commit
      (hourly) PushJournal()     — git push with rebase + retry

fw dashboard
  → LoadBlocks() × 6 periods (today/yesterday/week/month/year/all)
  → BuildStandup()               — cached standup for today/yesterday
  → buildDashPayload()           — assemble JSON
  → RenderDashboard()            — inject JSON into embedded HTML
  → open in browser
```

## Why a single embedded HTML file?

See [ADR-0003](adr/0003-static-dashboard-embed.md).
