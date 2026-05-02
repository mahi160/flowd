package ai_sessions

import (
	"encoding/json"
	"time"
)

type PiSession struct {
	Type      string           `json:"type"`
	Version   int              `json:"version"`
	Id        string           `json:"id"`
	Timestamp string           `json:"timestamp"`
	Cwd       string           `json:"cwd"`
	Message   *PiMessageData   `json:"message,omitempty"`
}

type PiMessageData struct {
	Usage PiUsageData `json:"usage,omitempty"`
}

type PiUsageData struct {
	Input  int     `json:"input"`
	Output int     `json:"output"`
	Cost   float64 `json:"total"`
}

type PiProcessor struct{}

func (p *PiProcessor) Name() string { return "pi" }

func (p *PiProcessor) ProcessEntry(data []byte) (*SessionStats, error) {
	var entry PiSession
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	// Only process message entries with usage data
	if entry.Type != "message" || entry.Message == nil {
		return nil, nil
	}

	// Parse timestamp
	ts, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		ts = time.Now()
	}

	// Extract project name from cwd, remove leading /Users/mahi/
	project := entry.Cwd
	if len(project) > 50 {
		project = project[len(project)-50:]
	}

	return &SessionStats{
		Tool:        "pi",
		Project:     project,
		SessionID:   entry.Id,
		Timestamp:   ts,
		TokensRead:  entry.Message.Usage.Input,
		TokensWrite: entry.Message.Usage.Output,
		TokensCache: 0, // pi doesn't have cache metrics in this format
		Cost:        entry.Message.Usage.Cost,
	}, nil
}

func init() {
	Register(&PiProcessor{})
}
