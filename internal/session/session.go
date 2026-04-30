package session

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/mahi/flowd/internal/collector/process"
	"github.com/mahi/flowd/internal/collector/tmux"
	"github.com/mahi/flowd/internal/db"
)

type EventType string

const (
	EventPaneActive EventType = "pane_active"
	EventCwdChange  EventType = "cwd_change"
	EventCmdChange  EventType = "cmd_change"
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
	db        *db.DB
	interval  time.Duration
	watchDirs []string
	last      *tmux.PaneState
}

func NewTracker(d *db.DB, pollSec int, watchDirs []string) *Tracker {
	return &Tracker{
		db:        d,
		interval:  time.Duration(pollSec) * time.Second,
		watchDirs: watchDirs,
	}
}

func (t *Tracker) Run(ctx context.Context) {
	// Wait for tmux to come up — useful when started from tmux.conf before server is ready
	t.waitForTmux(ctx)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !tmux.IsRunning() {
				slog.Debug("tmux gone, waiting")
				t.waitForTmux(ctx)
				continue
			}
			t.poll()
		}
	}
}

// waitForTmux blocks until tmux is available or ctx is cancelled.
func (t *Tracker) waitForTmux(ctx context.Context) {
	if tmux.IsRunning() {
		return
	}
	slog.Info("waiting for tmux to start")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			if tmux.IsRunning() {
				slog.Info("tmux detected, starting collector")
				return
			}
		}
	}
}

func (t *Tracker) poll() {
	// Track the most-recently-active pane across ALL sessions globally
	state, err := tmux.MostActive()
	if err != nil {
		slog.Debug("tmux poll failed", "err", err)
		return
	}

	if !t.inWatchDirs(state.Cwd) {
		slog.Debug("cwd outside watch_dirs, skipping", "cwd", state.Cwd)
		return
	}

	repo := process.RepoName(state.Cwd)
	cat := process.ClassifyCommand(state.Command)

	meta := PaneMeta{
		Session:  state.Session,
		Window:   state.Window,
		Pane:     state.Pane,
		Command:  state.Command,
		Category: cat,
		Cwd:      state.Cwd,
		Repo:     repo,
	}

	metaJSON, _ := json.Marshal(meta)

	t.writeEvent(EventPaneActive, state.PaneID, string(metaJSON))

	if t.last != nil {
		if t.last.Cwd != state.Cwd {
			t.writeEvent(EventCwdChange, state.Cwd, string(metaJSON))
		}
		if t.last.Command != state.Command {
			t.writeEvent(EventCmdChange, state.Command, string(metaJSON))
		}
	}

	t.last = state
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

func (t *Tracker) writeEvent(typ EventType, value, meta string) {
	_, err := t.db.Exec(
		`INSERT INTO events (ts, type, value, meta) VALUES (?, ?, ?, ?)`,
		time.Now().UTC(), string(typ), value, meta,
	)
	if err != nil {
		slog.Error("db write event", "err", err)
	}
}
