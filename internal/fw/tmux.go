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
	PanePID int
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

// parsePane parses a single line of tmux display-message output.
// Expected format (7 pipe-separated fields):
//
//	#{session_name}|#{window_name}|#{pane_index}|#{pane_id}|#{pane_pid}|#{pane_current_command}|#{pane_current_path}
func parsePane(raw string) (*Pane, error) {
	parts := strings.SplitN(strings.TrimSpace(raw), "|", 7)
	if len(parts) < 7 {
		return nil, fmt.Errorf("bad tmux output: %q", raw)
	}
	var panePID int
	fmt.Sscan(parts[4], &panePID)
	return &Pane{
		Session: parts[0], Window: parts[1], Pane: parts[2],
		PaneID: parts[3], PanePID: panePID,
		Command: parts[5], Cwd: parts[6],
	}, nil
}

// ActivePane returns the active pane of the given session.
func ActivePane(session string) (*Pane, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "-t", session,
		"-F", "#{session_name}|#{window_name}|#{pane_index}|#{pane_id}|#{pane_pid}|#{pane_current_command}|#{pane_current_path}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("tmux display-message: %w", err)
	}
	return parsePane(string(out))
}

