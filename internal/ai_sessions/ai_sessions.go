package ai_sessions

import (
	"time"
)

type SessionStats struct {
	Tool        string
	Project     string
	SessionID   string
	Timestamp   time.Time
	TokensRead  int
	TokensWrite int
	TokensCache int
	Cost        float64
	ToolsCalled int
	FilesChanged int
}

type Processor interface {
	Name() string
	ProcessEntry(data []byte) (*SessionStats, error)
}

var processors = make(map[string]Processor)

func Register(p Processor) {
	processors[p.Name()] = p
}

func GetProcessors() map[string]Processor {
	return processors
}
