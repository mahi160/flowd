package ai_sessions

import (
	"encoding/json"
	"path/filepath"
	"time"
)

type ClaudeEntry struct {
	Type      string             `json:"type"`
	Timestamp string             `json:"timestamp"`
	SessionID string             `json:"sessionId"`
	Cwd       string             `json:"cwd"`
	Message   *ClaudeMessageData `json:"message"`
}

type ClaudeMessageData struct {
	Model string          `json:"model"`
	Usage ClaudeUsageData `json:"usage"`
}

type ClaudeUsageData struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

type ClaudeProcessor struct{}

func (p *ClaudeProcessor) Name() string { return "claude-code" }

// FileSessionID: Claude filenames are just the UUID (e.g.
// "67849373-75bc-42c6-98ac-b2e061ffdf15"), so the base is the session ID.
func (p *ClaudeProcessor) FileSessionID(base string) string {
	return base
}

// SessionCwd: Claude stores cwd on every assistant entry, so no session-level extraction needed.
func (p *ClaudeProcessor) SessionCwd(_ []byte) string { return "" }

func (p *ClaudeProcessor) ProcessEntry(data []byte, fileSessionID, sessionCwd string) (*SessionStats, error) {
	var entry ClaudeEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if entry.Type != "assistant" || entry.Timestamp == "" || entry.Message == nil {
		return nil, nil
	}

	ts, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		ts = time.Now()
	}

	project := filepath.Base(entry.Cwd)
	if project == "." || project == "" {
		project = entry.Cwd
	}

	// Use file-level session ID; fall back to entry's own sessionId field.
	sid := fileSessionID
	if sid == "" {
		sid = entry.SessionID
	}

	return &SessionStats{
		Tool:        "claude-code",
		Project:     project,
		SessionID:   sid,
		Model:       entry.Message.Model,
		Timestamp:   ts,
		TokensRead:  entry.Message.Usage.InputTokens + entry.Message.Usage.CacheReadInputTokens,
		TokensWrite: entry.Message.Usage.OutputTokens,
		TokensCache: entry.Message.Usage.CacheCreationInputTokens + entry.Message.Usage.CacheReadInputTokens,
		Cost:        0, // Claude CLI doesn't expose cost in session files
	}, nil
}

func init() {
	Register(&ClaudeProcessor{})
}
