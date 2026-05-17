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

// classifyCommand maps a process name to a tool category used for
// language inference. The raw command name is stored in ByTool;
// the category is an internal grouping only.
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
}

type Tracker struct {
	db           *DB
	interval     time.Duration
	idleSec      int
	watchDirs    []string
	last         *Pane
	repoCache    map[string]string // cwd → repo name (avoids a git subprocess per poll)
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
			if !TmuxRunning() {
				slog.Debug("tmux gone")
				t.last = nil
				t.waitForTmux(ctx)
				continue
			}
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
	session, idle := AttachedSession()
	if session == "" {
		if t.last != nil {
			slog.Debug("tmux detached")
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
	cat := classifyCommand(p.Command)
	pl := GetPlatform()
	meta, _ := json.Marshal(PaneMeta{
		Session: p.Session, Window: p.Window, Pane: p.Pane,
		Command: p.Command, Category: cat, Cwd: p.Cwd, Repo: repo,
		Machine: pl.Machine, OS: pl.OS,
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
