package ai_sessions

import (
	"bufio"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
)

type Service struct {
	db    *sql.DB
	paths map[string]string // tool name -> path
}

// NewService initializes and runs an immediate sync
func NewService(db *sql.DB, paths map[string]string) *Service {
	s := &Service{db: db, paths: paths}
	// Initial sync on startup
	if err := s.RunSync(); err != nil {
		slog.Error("initial ai_sessions sync", "err", err)
	}
	return s
}

func (s *Service) RunSync() error {
	for tool, dir := range s.paths {
		proc, ok := processors[tool]
		if !ok {
			slog.Warn("no processor found for tool", "tool", tool)
			continue
		}
		if err := s.processDirectory(tool, dir, proc); err != nil {
			slog.Error("error processing directory", "tool", tool, "dir", dir, "err", err)
		}
	}
	return nil
}

func (s *Service) processDirectory(tool, dir string, proc Processor) error {
	// 1. Get watermark
	var lastProcessed int64
	err := s.db.QueryRow(`SELECT offset FROM ai_sessions_watermark WHERE tool = ?`, tool).Scan(&lastProcessed)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 2. Find files recursively (not just at top level)
	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() && filepath.Ext(path) == ".jsonl" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		slog.Warn("walk directory", "dir", dir, "err", err)
		return err
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			slog.Error("open file", "file", file, "err", err)
			continue
		}
		defer f.Close()

		// Simplified: read and process new entries
		scanner := bufio.NewScanner(f)
		var currentOffset int64
		for scanner.Scan() {
			currentOffset++
			if currentOffset <= lastProcessed {
				continue
			}

			// Process
			stats, err := proc.ProcessEntry(scanner.Bytes())
			if err != nil || stats == nil {
				continue
			}
			
			// Save to DB
			s.saveStats(stats)
		}

		// Update watermark (simplified to per-tool, ideally per-file)
		_, err = s.db.Exec(`INSERT INTO ai_sessions_watermark (tool, offset) VALUES (?, ?) ON CONFLICT(tool) DO UPDATE SET offset=excluded.offset`, tool, currentOffset)
	}
	return nil
}

func (s *Service) saveStats(stats *SessionStats) {
	_, err := s.db.Exec(`INSERT INTO ai_sessions_raw (tool, project, session_id, ts, tokens_read, tokens_write, tokens_cache, cost, tools_called, files_changed) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stats.Tool, stats.Project, stats.SessionID, stats.Timestamp, stats.TokensRead, stats.TokensWrite, stats.TokensCache, stats.Cost, stats.ToolsCalled, stats.FilesChanged)
	if err != nil {
		slog.Error("save stats", "err", err)
	}
}
