package fw

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"
)

const (
	EvActive        = "pane_active"
	EvCwdChange     = "cwd_change"
	EvSessionChange = "session_change" // project/context switch
)

// pidCacheEntry holds a resolved AI-CLI result for a specific pane PID
// so we don't call ResolveAICLI on every 3-second poll tick.
type pidCacheEntry struct {
	tool      string // "" means "checked, not an AI CLI"
	expiresAt time.Time
}

const pidCacheTTL = 5 * time.Second

// nvimCacheEntry caches the most recently read filetype for a nvim pane CWD
// so we don't scan the state directory on every 3-second poll tick.
type nvimCacheEntry struct {
	filetype  string
	expiresAt time.Time
}

// nvimCacheTTL is shorter than nvimStateTTL (15s) so the cached value is
// refreshed well before the state file itself would go stale.
const nvimCacheTTL = 5 * time.Second

// classifyCommand maps a process name to a tool category used for
// language inference. The raw command name is stored in ByTool;
// the category is an internal grouping only.
//
// Note: the "runtime" list here overlaps with (but is distinct from) the
// interpreters map in proctree.go — that one answers "could an AI CLI be
// hiding beneath this process?", this one answers "what category is it?".
// Keep both in mind when adding a new runtime.
func classifyCommand(cmd string) string {
	switch cmd {
	case "nvim", "vim", "vi", "nano", "emacs", "hx", "micro", "code", "subl", "gedit", "kate":
		return "editor"
	case "claude", "pi", "aider", "codex", "opencode", "llm", "sgpt", "tgpt", "gemini", "copilot":
		return "ai"
	case "git", "lazygit", "tig", "gh", "gitu":
		return "git"
	case "zsh", "bash", "fish", "sh", "dash", "nu", "elvish":
		return "shell"
	case "node", "bun", "deno", "python", "python3", "go", "cargo", "ruby", "java", "php", "elixir", "erl", "iex", "ghci", "julia", "swift", "dotnet", "lua":
		return "runtime"
	case "docker", "docker-compose", "kubectl", "k9s", "podman", "helm":
		return "docker"
	case "psql", "mysql", "sqlite3", "redis-cli", "mongosh", "pgcli", "mycli":
		return "db"
	}
	return "other"
}

type PaneMeta struct {
	Session  string `json:"session"`
	Window   string `json:"window"`
	Pane     string `json:"pane"`
	Command  string `json:"command"`
	Category string `json:"category"`
	Cwd      string `json:"cwd"`
	Repo     string `json:"repo,omitempty"`
	Machine  string `json:"machine,omitempty"`
	OS       string `json:"os,omitempty"`
	// NvimFiletype is populated by the flowd.lua plugin when the pane is nvim.
	// Empty string means plugin absent or not an nvim pane.
	NvimFiletype string `json:"nvim_ft,omitempty"`
}

type Tracker struct {
	db        *DB
	interval  time.Duration
	idleSec   int
	watchDirs []string
	last      *Pane
	// repoCache avoids a git subprocess on every poll tick.
	repoCache map[string]string // cwd → repo name
	// pidCache caches AI-CLI resolution per pane PID to avoid calling
	// ResolveAICLI on every 3-second tick (only used for interpreter panes).
	pidCache map[int]pidCacheEntry
	// nvimCache caches the nvim plugin filetype per pane CWD to avoid
	// scanning the state directory on every poll tick.
	nvimCache map[string]nvimCacheEntry
	pollCount int // used to schedule periodic cache cleanup
}

func NewTracker(d *DB, pollSec, idleSec int, watchDirs []string) *Tracker {
	dirs := make([]string, len(watchDirs))
	for i, x := range watchDirs {
		dirs[i] = strings.TrimRight(x, "/") + "/"
	}
	return &Tracker{
		db:        d,
		interval:  time.Duration(pollSec) * time.Second,
		idleSec:   idleSec,
		watchDirs: dirs,
		repoCache: map[string]string{},
		pidCache:  map[int]pidCacheEntry{},
		nvimCache: map[string]nvimCacheEntry{},
	}
}

func (t *Tracker) Run(ctx context.Context) {
	t.waitForTmux(ctx)
	tk := time.NewTicker(t.interval)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			// No separate TmuxRunning() check per tick: AttachedSession()
			// inside poll() already returns "" when the server is gone,
			// saving one tmux subprocess every interval.
			t.poll()
		}
	}
}

func (t *Tracker) waitForTmux(ctx context.Context) {
	for !TmuxRunning() {
		slog.Info("waiting for tmux")
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (t *Tracker) poll() {
	t.pollCount++
	// Trim stale pidCache entries every 300 polls (~15 min at default 3s interval)
	// so panes that closed long ago don't accumulate indefinitely.
	if t.pollCount%300 == 0 {
		now := time.Now()
		for pid, entry := range t.pidCache {
			if now.After(entry.expiresAt) {
				delete(t.pidCache, pid)
			}
		}
		for cwd, entry := range t.nvimCache {
			if now.After(entry.expiresAt) {
				delete(t.nvimCache, cwd)
			}
		}
	}

	session, idle := AttachedSession()
	if session == "" {
		if t.last != nil {
			slog.Debug("tmux detached or gone")
			t.last = nil
		}
		return
	}
	// Checked after the session so ioreg (macOS) only runs while attached.
	if ScreenClosed() {
		if t.last != nil {
			slog.Debug("screen closed, pausing tracking")
			t.last = nil
		}
		return
	}
	if t.idleSec > 0 && idle >= t.idleSec {
		if t.last != nil {
			slog.Debug("user idle", "sec", idle)
			t.last = nil
		}
		return
	}
	p, err := ActivePane(session)
	if err != nil {
		slog.Debug("active pane", "err", err)
		return
	}
	if !t.inWatchDirs(p.Cwd) {
		slog.Debug("cwd outside watch_dirs", "cwd", p.Cwd)
		return
	}

	// Cache repo name per cwd to avoid a git subprocess on every poll tick.
	repo, ok := t.repoCache[p.Cwd]
	if !ok {
		repo = RepoName(p.Cwd)
		t.repoCache[p.Cwd] = repo
	}

	cmd, cat := t.resolveCommand(p)

	// When the pane is nvim, check for the plugin state file so the block
	// builder can use the real filetype for language attribution.
	// Match by cwd, NOT by pane PID: tmux #{pane_pid} is the shell's PID;
	// the plugin names its file after nvim's own PID (different process).
	var nvimFT string
	if cmd == "nvim" {
		nvimFT = t.resolveNvimFiletype(p.Cwd)
	}

	pl := GetPlatform()
	meta, _ := json.Marshal(PaneMeta{
		Session: p.Session, Window: p.Window, Pane: p.Pane,
		Command: cmd, Category: cat, Cwd: p.Cwd, Repo: repo,
		Machine: pl.Machine, OS: pl.OS,
		NvimFiletype: nvimFT,
	})

	t.write(EvActive, p.PaneID, string(meta))
	if t.last != nil {
		if t.last.Cwd != p.Cwd {
			t.write(EvCwdChange, p.Cwd, string(meta))
		}
		if t.last.Session != p.Session {
			t.write(EvSessionChange, p.Session, string(meta))
		}
	}
	t.last = p
}

// resolveCommand returns the effective (command, category) pair for a pane.
//
// Fast path: if the pane command is not a known interpreter, classify it
// directly — zero overhead.
//
// Slow path: when the command IS a known interpreter (node, python, …), the
// actual tool running inside the pane may be an AI CLI whose process is a
// child of the shell (e.g. tmux sees "node" but the real foreground process
// is "pi"). We walk the process tree once and cache the result per pane PID
// for pidCacheTTL to keep the common case cheap.
func (t *Tracker) resolveCommand(p *Pane) (cmd, cat string) {
	cmd = p.Command
	cat = classifyCommand(cmd)

	// Fast path: not an interpreter → nothing to resolve.
	if cat != "runtime" || !IsInterpreter(cmd) {
		return cmd, cat
	}

	// Check cache first.
	now := time.Now()
	if entry, ok := t.pidCache[p.PanePID]; ok && now.Before(entry.expiresAt) {
		if entry.tool != "" {
			return entry.tool, "ai"
		}
		return cmd, cat // cached "not an AI CLI"
	}

	// Walk the process tree.
	tool := ResolveAICLI(p.PanePID)
	t.pidCache[p.PanePID] = pidCacheEntry{
		tool:      tool,
		expiresAt: now.Add(pidCacheTTL),
	}
	if tool != "" {
		slog.Debug("resolved interpreter to AI CLI",
			"pane_pid", p.PanePID, "interpreter", cmd, "tool", tool)
		return tool, "ai"
	}
	return cmd, cat
}

// resolveNvimFiletype returns the filetype for a nvim pane, using a short-lived
// cache to avoid scanning the state directory on every 3-second poll tick.
func (t *Tracker) resolveNvimFiletype(cwd string) string {
	now := time.Now()
	if e, ok := t.nvimCache[cwd]; ok && now.Before(e.expiresAt) {
		return e.filetype
	}
	ft := NvimFiletype(cwd)
	t.nvimCache[cwd] = nvimCacheEntry{filetype: ft, expiresAt: now.Add(nvimCacheTTL)}
	return ft
}

func (t *Tracker) inWatchDirs(cwd string) bool {
	if len(t.watchDirs) == 0 {
		return true
	}
	cwd = strings.TrimRight(cwd, "/") + "/"
	for _, d := range t.watchDirs {
		if strings.HasPrefix(cwd, d) {
			return true
		}
	}
	return false
}

func (t *Tracker) write(typ, value, meta string) {
	_, err := t.db.Exec(
		`INSERT INTO events (ts, type, value, meta) VALUES (?, ?, ?, ?)`,
		time.Now().UTC(), typ, value, meta,
	)
	if err != nil {
		slog.Error("event write", "err", err)
	}
}
