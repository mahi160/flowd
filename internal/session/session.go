package session

import (
	"context"
	"encoding/json"
	"log/slog"
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
	db       *db.DB
	interval time.Duration
	last     *tmux.PaneState
}

func NewTracker(d *db.DB, pollSec int) *Tracker {
	return &Tracker{
		db:       d,
		interval: time.Duration(pollSec) * time.Second,
	}
}

func (t *Tracker) Run(ctx context.Context) {
	if !tmux.IsRunning() {
		slog.Warn("tmux not running, collector idle")
	}

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.poll()
		}
	}
}

func (t *Tracker) poll() {
	state, err := tmux.Active()
	if err != nil {
		slog.Debug("tmux poll failed", "err", err)
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

	// Always record pane_active tick
	t.writeEvent(EventPaneActive, state.PaneID, string(metaJSON))

	// Record diffs only
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

func (t *Tracker) writeEvent(typ EventType, value, meta string) {
	_, err := t.db.Exec(
		`INSERT INTO events (ts, type, value, meta) VALUES (?, ?, ?, ?)`,
		time.Now().UTC(), string(typ), value, meta,
	)
	if err != nil {
		slog.Error("db write event", "err", err)
	}
}
