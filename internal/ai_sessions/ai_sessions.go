package ai_sessions

import (
	"time"
)

// SessionStats is a single message-level data point from an AI tool.
type SessionStats struct {
	Tool         string
	Project      string
	SessionID    string
	Model        string
	Timestamp    time.Time
	TokensRead   int
	TokensWrite  int
	TokensCache  int
	Cost         float64
	ToolsCalled  int
	FilesChanged int
}

// AggregatedSession is a session-level summary built from multiple messages.
type AggregatedSession struct {
	Tool         string  `json:"tool"`
	Project      string  `json:"project"`
	SessionID    string  `json:"session_id"`
	Model        string  `json:"model"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	StartUnix    int64   `json:"start_unix"` // for deterministic sort on the backend
	TotalInput   int     `json:"total_input"`
	TotalOutput  int     `json:"total_output"`
	TotalCache   int     `json:"total_cache"`
	TotalCost    float64 `json:"total_cost"`
	MessageCount int     `json:"message_count"`
}

// ToolSummary is the per-tool aggregate sent to the dashboard.
type ToolSummary struct {
	Tool           string             `json:"tool"`
	TotalCost      float64            `json:"total_cost"`
	TotalInput     int                `json:"total_input"`
	TotalOutput    int                `json:"total_output"`
	TotalCache     int                `json:"total_cache"`
	SessionCount   int                `json:"session_count"`
	MessageCount   int                `json:"message_count"`
	TopModel       string             `json:"top_model"`
	ModelBreakdown map[string]int     `json:"model_breakdown"`
	Sessions       []AggregatedSession `json:"sessions"`
}

// Processor handles a single AI tool's JSONL session files.
type Processor interface {
	// Name returns the tool key used in config (e.g. "pi", "claude-code").
	Name() string

	// FileSessionID extracts the session ID from the filename base (without
	// the .jsonl extension). Called once per file before processing its lines.
	FileSessionID(base string) string

	// SessionCwd extracts the working directory from the first line of a session
	// file (the session metadata entry). Returns "" if the tool stores cwd
	// per-message instead of once per session.
	SessionCwd(firstLine []byte) string

	// ProcessEntry parses one JSONL line. fileSessionID is the session ID
	// derived from the filename; sessionCwd is from SessionCwd (may be empty
	// if the tool embeds cwd in each message). Returns nil, nil to skip.
	ProcessEntry(data []byte, fileSessionID, sessionCwd string) (*SessionStats, error)
}

var processors = make(map[string]Processor)

func Register(p Processor) {
	processors[p.Name()] = p
}

func GetProcessors() map[string]Processor {
	return processors
}
