package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

type PaneState struct {
	Session  string
	Window   string
	Pane     string
	Command  string
	Cwd      string
	PaneID   string
	WindowID string
}

// Active returns the current active pane state by querying tmux.
func Active() (*PaneState, error) {
	// #{session_name}|#{window_name}|#{pane_index}|#{pane_current_command}|#{pane_current_path}|#{pane_id}|#{window_id}
	out, err := exec.Command("tmux", "display-message", "-p",
		"#{session_name}|#{window_name}|#{pane_index}|#{pane_current_command}|#{pane_current_path}|#{pane_id}|#{window_id}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("tmux unavailable: %w", err)
	}

	parts := strings.SplitN(strings.TrimSpace(string(out)), "|", 7)
	if len(parts) < 7 {
		return nil, fmt.Errorf("unexpected tmux output: %q", string(out))
	}

	return &PaneState{
		Session:  parts[0],
		Window:   parts[1],
		Pane:     parts[2],
		Command:  parts[3],
		Cwd:      parts[4],
		PaneID:   parts[5],
		WindowID: parts[6],
	}, nil
}

// IsRunning checks whether a tmux server is running.
func IsRunning() bool {
	err := exec.Command("tmux", "list-sessions").Run()
	return err == nil
}

// Sessions returns all session names.
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
