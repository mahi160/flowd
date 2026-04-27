package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

type PaneState struct {
	Session    string
	Window     string
	Pane       string
	Command    string
	Cwd        string
	PaneID     string
	WindowID   string
	LastActive int64 // activity epoch from tmux, used to pick most-recently-active pane
}

// AllPanes returns all panes across all sessions, globally.
func AllPanes() ([]PaneState, error) {
	// list-panes -a covers every session on the server
	out, err := exec.Command("tmux", "list-panes", "-a", "-F",
		"#{session_name}|#{window_name}|#{pane_index}|#{pane_current_command}|#{pane_current_path}|#{pane_id}|#{window_id}|#{pane_last_used}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("tmux unavailable: %w", err)
	}

	var panes []PaneState
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 8)
		if len(parts) < 8 {
			continue
		}
		var lastActive int64
		fmt.Sscan(parts[7], &lastActive)
		panes = append(panes, PaneState{
			Session:    parts[0],
			Window:     parts[1],
			Pane:       parts[2],
			Command:    parts[3],
			Cwd:        parts[4],
			PaneID:     parts[5],
			WindowID:   parts[6],
			LastActive: lastActive,
		})
	}
	return panes, nil
}

// MostActive returns the pane with the most recent activity across all sessions.
func MostActive() (*PaneState, error) {
	panes, err := AllPanes()
	if err != nil {
		return nil, err
	}
	if len(panes) == 0 {
		return nil, fmt.Errorf("no panes found")
	}
	best := &panes[0]
	for i := range panes[1:] {
		if panes[i+1].LastActive > best.LastActive {
			best = &panes[i+1]
		}
	}
	return best, nil
}

// IsRunning checks whether a tmux server is running.
func IsRunning() bool {
	return exec.Command("tmux", "list-sessions").Run() == nil
}

// WaitUntilRunning blocks until tmux is available, checking every checkSec seconds.
// Returns when tmux is up or ctx is cancelled.
func Sessions() ([]string, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	var sessions []string
	for _, s := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if s != "" {
			sessions = append(sessions, s)
		}
	}
	return sessions, nil
}
