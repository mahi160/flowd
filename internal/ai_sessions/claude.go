package ai_sessions

import (
	"encoding/json"
	"time"
)

type ClaudeEntry struct {
	Type      string       `json:"type"`
	Timestamp string       `json:"timestamp"`
	SessionID string       `json:"sessionId"`
	Cwd       string       `json:"cwd"`
	Message   *MessageData `json:"message"`
}

type MessageData struct {
	Usage UsageData `json:"usage"`
}

type UsageData struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

type ClaudeProcessor struct{}

func (p *ClaudeProcessor) Name() string { return "claude-code" }

func (p *ClaudeProcessor) ProcessEntry(data []byte) (*SessionStats, error) {
	var entry ClaudeEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	// Only process entries with message and usage data
	if entry.Type != "assistant" || entry.Timestamp == "" || entry.Message == nil {
		return nil, nil
	}

	// Parse timestamp
	ts, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		ts = time.Now()
	}

	// Extract project name from cwd
	project := entry.Cwd
	if len(project) > 50 {
		project = project[len(project)-50:]
	}

	return &SessionStats{
		Tool:        "claude-code",
		Project:     project,
		SessionID:   entry.SessionID,
		Timestamp:   ts,
		TokensRead:  entry.Message.Usage.InputTokens,
		TokensWrite: entry.Message.Usage.OutputTokens,
		TokensCache: entry.Message.Usage.CacheReadInputTokens,
		Cost:        0, // TODO: Calculate cost based on token counts
	}, nil
}

func init() {
	Register(&ClaudeProcessor{})
}
