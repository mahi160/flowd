package ai_sessions

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"time"
)

type PiSession struct {
	Type      string         `json:"type"`
	Id        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Cwd       string         `json:"cwd"`
	Message   *PiMessageData `json:"message,omitempty"`
}

type PiMessageData struct {
	Model string      `json:"model"`
	Usage PiUsageData `json:"usage,omitempty"`
}

type PiUsageData struct {
	Input      int        `json:"input"`
	Output     int        `json:"output"`
	CacheRead  int        `json:"cacheRead"`
	CacheWrite int        `json:"cacheWrite"`
	Cost       PiCostData `json:"cost"`
}

type PiCostData struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
	Total      float64 `json:"total"`
}

type PiProcessor struct{}

func (p *PiProcessor) Name() string { return "pi" }

// FileSessionID extracts the session ID from the pi filename format:
// "2026-05-17T11-27-05-880Z_019e35b0-9698-7089-be4c-c4c510057df7" → last segment after "_".
func (p *PiProcessor) FileSessionID(base string) string {
	if idx := strings.LastIndex(base, "_"); idx >= 0 {
		return base[idx+1:]
	}
	return base
}

// SessionCwd extracts the cwd from pi's session metadata line (type:"session").
// Pi only stores cwd once, in the first line of each file.
func (p *PiProcessor) SessionCwd(firstLine []byte) string {
	var entry PiSession
	if err := json.Unmarshal(firstLine, &entry); err != nil {
		return ""
	}
	if entry.Type != "session" {
		return ""
	}
	return entry.Cwd
}

func (p *PiProcessor) ProcessEntry(data []byte, fileSessionID, sessionCwd string) (*SessionStats, error) {
	var entry PiSession
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if entry.Type != "message" || entry.Message == nil {
		return nil, nil
	}

	ts, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		ts = time.Now()
	}

	// Use sessionCwd (from first line) since message entries have empty cwd.
	cwd := sessionCwd
	if cwd == "" {
		cwd = entry.Cwd
	}
	project := filepath.Base(cwd)
	if project == "." || project == "" {
		project = cwd
	}

	return &SessionStats{
		Tool:        "pi",
		Project:     project,
		SessionID:   fileSessionID,
		Model:       entry.Message.Model,
		Timestamp:   ts,
		TokensRead:  entry.Message.Usage.Input,
		TokensWrite: entry.Message.Usage.Output,
		TokensCache: entry.Message.Usage.CacheRead + entry.Message.Usage.CacheWrite,
		Cost:        entry.Message.Usage.Cost.Total,
	}, nil
}

func init() {
	Register(&PiProcessor{})
}
