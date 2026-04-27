package fw

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Pane struct {
	Session string
	Window  string
	Pane    string
	PaneID  string
	Command string
	Cwd     string
}

func TmuxRunning() bool {
	return exec.Command("tmux", "list-sessions").Run() == nil
}

// AttachedSession returns the session name of the most-recently-active
// attached client, plus the seconds since that client last had input.
// Returns ("", 0) if no client is attached.
func AttachedSession() (string, int) {
	out, err := exec.Command("tmux", "list-clients", "-F", "#{client_session}|#{client_activity}").Output()
	if err != nil {
		return "", 0
	}
	now := time.Now().Unix()
	bestSession := ""
	bestIdle := int(1 << 30)
	for _, ln := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		parts := strings.SplitN(ln, "|", 2)
		if len(parts) != 2 {
			continue
		}
		var act int64
		fmt.Sscan(parts[1], &act)
		idle := int(now - act)
		if idle < 0 {
			idle = 0
		}
		if idle < bestIdle {
			bestIdle, bestSession = idle, parts[0]
		}
	}
	if bestSession == "" {
		return "", 0
	}
	return bestSession, bestIdle
}

// ActivePane returns the active pane of the given session.
func ActivePane(session string) (*Pane, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "-t", session,
		"-F", "#{session_name}|#{window_name}|#{pane_index}|#{pane_id}|#{pane_current_command}|#{pane_current_path}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("tmux display-message: %w", err)
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), "|", 6)
	if len(parts) < 6 {
		return nil, fmt.Errorf("bad tmux output: %q", out)
	}
	return &Pane{
		Session: parts[0], Window: parts[1], Pane: parts[2],
		PaneID: parts[3], Command: parts[4], Cwd: parts[5],
	}, nil
}

