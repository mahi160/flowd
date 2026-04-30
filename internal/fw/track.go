package fw

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"
)

const (
	EvActive    = "pane_active"
	EvCwdChange = "cwd_change"
	EvCmdChange = "cmd_change"
)

type PaneMeta struct {
	Session  string `json:"session"`
	Window   string `json:"window"`
	Pane     string `json:"pane"`
	Command  string `json:"command"`
	Category string `json:"category"`
	Cwd      string `json:"cwd"`
	Repo     string `json:"repo,omitempty"`
}

type Tracker struct {
	db        *DB
	interval  time.Duration
	watchDirs []string
	last      *Pane
}

func NewTracker(d *DB, pollSec int, watchDirs []string) *Tracker {
	return &Tracker{
		db:        d,
		interval:  time.Duration(pollSec) * time.Second,
		watchDirs: watchDirs,
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
	session := AttachedSession()
	if session == "" {
		// detached / out of focus — do not record
		if t.last != nil {
			slog.Debug("tmux detached, idle")
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

	repo := RepoName(p.Cwd)
	cat := ClassifyCommand(p.Command)
	meta, _ := json.Marshal(PaneMeta{
		Session: p.Session, Window: p.Window, Pane: p.Pane,
		Command: p.Command, Category: cat, Cwd: p.Cwd, Repo: repo,
	})

	t.write(EvActive, p.PaneID, string(meta))
	if t.last != nil {
		if t.last.Cwd != p.Cwd {
			t.write(EvCwdChange, p.Cwd, string(meta))
		}
		if t.last.Command != p.Command {
			t.write(EvCmdChange, p.Command, string(meta))
		}
	}
	t.last = p
}

func (t *Tracker) inWatchDirs(cwd string) bool {
	if len(t.watchDirs) == 0 {
		return true
	}
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
