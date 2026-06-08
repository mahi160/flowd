# flowd — refactor & improvement plan

## Context

The user asked to "rewrite the whole thing" because they (a) don't understand how it
works, (b) suspect Linux is unsupported, (c) see AI CLIs mislabelled ("pi shows as
node"), (d) can't get rich data out of neovim, and (e) find the dashboard generic and
disorganized — and wondered whether to adopt Svelte 5.

**Key finding from exploration:** the codebase is already cleanly modular and the
dashboard is *already* Svelte 5 + Tailwind 4 + Vite, compiled to a single embedded
HTML file. So several premises behind "rewrite everything" don't hold. This plan
proposes targeted fixes to the genuine problems plus documentation, rather than a
ground-up rewrite. **Confirmed with the user.** All decisions in Open Questions are settled.

## Current architecture (as-is)

- `cmd/fw/main.go` — CLI entrypoint (cobra).
- `internal/fw/` — daemon lifecycle, tracker (tmux polling), scheduler (block/commit/push
  cadences), block builder, journal writer, git ops, SQLite, config, report, init wizard,
  AI summary runner, dashboard renderer.
- `internal/ai_sessions/` — pluggable per-tool processors (pi, claude-code) that read JSONL
  transcripts; a Service walks the dirs and ingests into `ai_sessions_raw`.
- `dashboard/` — Svelte 5 + Tailwind 4 + Vite app, built to a single HTML file, embedded
  into the binary via `go:embed`.

### How tmux reading works (answers user's "not sure how")
- `TmuxRunning`, `AttachedSession` (most-recently-active client + idle secs),
  `ActivePane` (`tmux display-message` for cwd + `pane_current_command`).
- Tracker polls every `poll_interval_sec` (default 3s), filters by watch dirs, classifies
  the command, writes an `events` row. Scheduler converts event ticks → focused minutes.

## The real problems

1. **AI CLIs mislabelled as `node`/`python`.** `pane_current_command` returns the
   *foreground* process. pi/claude run on a node/python interpreter, so tmux reports the
   interpreter, not the tool — and it gets classified as a runtime (→ JS/Python language
   inference). Token/cost data is captured separately by `ai_sessions` (reads JSONL), but
   the tmux *time attribution* is wrong.
2. **neovim is opaque.** The pane shows `nvim` but flowd can't see the open file / language
   / LSP / project from inside the editor.
3. **Linux untested.** Code is mostly portable; only `ScreenClosed` is macOS-specific
   (returns false elsewhere). Needs validation, not a rewrite.
4. **Dashboard design.** Already Svelte 5; the complaint is visual design + component
   organization + data density, not the framework.
5. **No navigable docs.** No CONTEXT.md/ADRs/docs site existed.

## Approach

**Decision: targeted refactor + the four fixes + docs. No ground-up rewrite.** Keep the
working core (idle detection, block resumption, WAL checkpointing, multi-machine journals,
language math). Tighten module boundaries only where it helps the fixes land cleanly.

## Open questions

- [x] Full rewrite vs. targeted refactor? → **Targeted refactor + fixes + docs.**
- [x] AI CLI detection → **Process-tree walk + argv match for known AI CLIs only.** When
  `pane_current_command` is a known interpreter (node/python/bun/deno/ruby), walk from
  `#{pane_pid}` (Linux `/proc`, macOS single `ps -axo pid,ppid,command`), match argv against
  AI-CLI patterns, relabel to the tool and category `ai`. Fast-path skip for non-interpreters.
  This also kills the phantom JS/Python language inference. JSONL stays source of truth for
  tokens/cost.
- [x] neovim → **Small Lua plugin** writes `{cwd, file, filetype, project}` to a known path
  (e.g. `~/.local/share/flowd/nvim/<pane_pid>.json`) on `BufEnter`/`DirChanged`; daemon reads
  it for `nvim` panes. **Daemon degrades gracefully when the plugin is absent** (falls back to
  git-diff language inference).
- [x] Linux → **Nice-to-have / portable tier.** macOS is primary. Keep portable, add a
  `/proc`-based process-tree walk so the AI-CLI fix works on Linux, but **don't break Linux**
  and don't invest in logind/DPMS lid signals yet. Idle-threshold is the only pause signal
  on Linux.
- [x] Dashboard delivery → **Keep static single-file embed.** Focus on redesign + reorg.
- [x] Dashboard centerpiece → **Long-term heatmaps/trends (GitHub-graph energy).** Collapse
  the 5 near-identical period branches into one config-driven layout (the real "disorganized"
  smell).
- [x] AI standup → **today/yesterday only**, grouped by project, source = **git commit messages
  per project + flowd's tracked time/files per project**. Reuses `RunAI` (point `ai_command` at
  gemini-flash / haiku). Shown on the dashboard AND injected into the journal's daily roll-up.
  Generate once and reuse in both (cache; regenerate when the day's data changes).
- [x] Journal md → **daily roll-up at top of each day** (totals + AI standup), per-block entries
  below. Implies the journal writer **regenerates each day's section from the DB** rather than
  blind-appending (cleaner, keeps the roll-up current).
- [x] Docs → **Markdown `docs/` folder + `CONTEXT.md` glossary + ADRs.** No HTML build step.

## Files to modify / add

**Fix 1 — AI-CLI detection (Go)**
- `internal/fw/proctree.go` (new) — resolve the foreground tool from `#{pane_pid}`: Linux via
  `/proc/<pid>/stat`+`/proc/<pid>/cmdline`, macOS via one `ps -axo pid,ppid,command`. Returns the
  matched AI-CLI name or "".
- `internal/fw/tmux.go` — add `pane_pid` to the `display-message` format; carry it on `Pane`.
- `internal/fw/track.go` — in `poll()`, when `classifyCommand`==`runtime` and the cmd is a known
  interpreter, call the resolver; if it matches an AI CLI, override `Command`+`Category=ai`.
  Cache per `pane_pid` to keep the common path cheap.

**Fix 2 — neovim plugin**
- `plugin/flowd.lua` (new) — on `BufEnter`/`DirChanged` write `~/.local/share/flowd/nvim/<pid>.json`
  = `{cwd,file,filetype,project,ts}`; clean up on `VimLeave`.
- `internal/fw/nvim.go` (new) — read+freshness-check that file for an `nvim` pane; expose filetype.
- `internal/fw/track.go` / `block.go` — when pane is `nvim` and plugin data is fresh, prefer its
  filetype for language attribution; otherwise fall back to existing git-diff inference. Graceful
  no-op when absent.

**Fix 3 — Linux** (no new files) — ensure `proctree.go` has a `/proc` path; smoke-test build/run on
  Linux; leave `ScreenClosed` returning false there.

**Dashboard redesign + standup**
- `dashboard/src/App.svelte` — collapse the 5 near-identical period branches into ONE config-driven
  layout (a per-period section list); make long-term heatmaps the centerpiece.
- `dashboard/src/components/*` — new `Standup.svelte`; restructure for the heatmap-first hero; prune/
  merge redundant components; establish one visual identity (tokens in `lib/theme.ts`/`styles.css`).
- `internal/fw/standup.go` (new) — build today/yesterday standup input: git commit messages per repo
  (`git log --since`) + flowd per-project time/files from blocks; call `RunAI`; cache by day-hash.
- `internal/fw/commands.go` — replace/extend `runAIRecap` to feed the standup; put result in payload.
- `internal/fw/git.go` — add `RecentCommits(repo, since)` helper (git log) alongside `DiffStat`.

**Journal md**
- `internal/fw/journal.go` — rewrite `WriteJournal` to regenerate each day's section from DB blocks
  (roll-up header: totals + standup, then per-block entries) instead of blind append; richer README
  optional.

**Docs**
- `docs/` (new) — `architecture.md`, `how-it-works.md` (tmux polling → events → blocks → journal/git),
  `data-model.md` (SQLite schema), `nvim-plugin.md`, `configuration.md`.
- `CONTEXT.md` (started) — glossary. `docs/adr/` — see ADRs below.

## Reuse (existing code — don't reinvent)

- `RunAI(ctx, command, prompt, body)` in `internal/fw/ai.go` — generic stdin→stdout AI runner; point
  `ai_command` at gemini-flash / haiku. Standup reuses this.
- `runAIRecap` + `--ai-recap` flag in `commands.go` — existing aggregate-AI pattern; standup extends it.
- `DiffStat`, `RepoRoot`, `RepoName`, `CurrentBranch`, `LangFromCommand`, `ScanLangs`, `runGit`/`git`
  in `git.go`.
- `LoadBlocks`, `PeriodRange`, `TextReport` in `report.go`; `BuildBlock`/`distributeByLines` in `block.go`.
- `classifyCommand` map in `track.go` — extend, don't replace.
- `ai_sessions.Processor`/`Register` plugin pattern — already the clean extension point for AI tools.
- Dashboard build pipeline: Vite single-file → `go:embed static/dashboard.html` → `RenderDashboard`.

## Steps

- [ ] **proctree resolver** (`proctree.go`) with Linux `/proc` + macOS `ps`, AI-CLI argv matcher; unit test.
- [ ] Thread `pane_pid` through `tmux.go` → `Pane`; wire resolver into `track.go` with per-pid cache.
- [ ] Verify phantom JS/Python languages disappear once AI panes are categorized `ai` (block_test).
- [ ] nvim Lua plugin + `nvim.go` reader + graceful fallback in language attribution.
- [ ] `standup.go`: gather git commits + per-project flowd stats for today/yesterday, RunAI, cache by day.
- [ ] Wire standup into dashboard payload (replace `runAIRecap`) and journal daily roll-up.
- [ ] Rewrite `WriteJournal` to regenerate-day-section-from-DB with roll-up + per-block entries.
- [ ] Dashboard: collapse 5 period branches into config-driven layout; heatmap-first hero; `Standup.svelte`;
  unify visual identity; rebuild embedded HTML.
- [ ] Linux smoke test (build + run under tmux); confirm no regressions on macOS.
- [ ] Write `docs/` pages, finish `CONTEXT.md`, add 3 ADRs:
  - `0001-process-tree-ai-detection.md` — why we don't trust `pane_current_command` for AI CLIs.
  - `0002-nvim-plugin-file-contract.md` — plugin writes JSON to a known path vs `--remote-expr`.
  - `0003-static-dashboard-embed.md` — single-file embed vs serving live from the daemon.
- [ ] Update `README.md` (mention nvim plugin, standup, gemini/haiku config).

## Verification

- [ ] `go build ./... && go test ./...` (extend `tmux_test.go`/`block_test.go` for proctree + AI relabel).
- [ ] Manual: open pi/claude in a tmux pane → `fw summary` shows it as `pi`/`claude`, NOT `node`; no
  phantom JS/Python in Languages.
- [ ] Manual: edit a file in nvim with the plugin installed → language reflects the real filetype even
  before committing; uninstall plugin → still works via git-diff.
- [ ] `fw dashboard` → heatmap-first layout, single code path across periods, standup panel renders.
- [ ] Inspect a monthly journal `.md`: day section has roll-up (totals + standup) above per-block entries;
  re-running a block for the same day updates the roll-up in place (no duplication).
- [ ] Build + run on a Linux box under tmux; confirm tracking + AI-CLI detection work.
- [ ] `docs/` browses cleanly on GitHub; `CONTEXT.md` terms match the code.
