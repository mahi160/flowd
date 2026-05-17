package ai_sessions

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	db    *sql.DB
	paths map[string]string // tool name → session directory
}

// NewService creates a Service. Call RunSync() explicitly to ingest sessions.
func NewService(db *sql.DB, paths map[string]string) *Service {
	return &Service{db: db, paths: paths}
}

// RunSync processes all configured tool directories. It returns a joined error
// of any per-tool failures so callers can log or surface them.
func (s *Service) RunSync() error {
	var errs []error
	for tool, dir := range s.paths {
		proc, ok := processors[tool]
		if !ok {
			slog.Warn("no processor for tool", "tool", tool)
			continue
		}
		if err := s.processDirectory(tool, dir, proc); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", tool, err))
		}
	}
	return errors.Join(errs...)
}

func (s *Service) processDirectory(tool, dir string, proc Processor) error {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if !info.IsDir() && filepath.Ext(path) == ".jsonl" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", dir, err)
	}

	for _, file := range files {
		if err := s.processFile(tool, file, proc); err != nil {
			slog.Error("process file", "file", file, "err", err)
		}
	}
	return nil
}

// processFile processes a single .jsonl file. The watermark is keyed by
// tool + file path so each file advances independently.
func (s *Service) processFile(tool, file string, proc Processor) error {
	watermarkKey := tool + "\x00" + file

	var lastProcessed int64
	err := s.db.QueryRow(`SELECT offset FROM ai_sessions_watermark WHERE tool = ?`, watermarkKey).Scan(&lastProcessed)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Extract the session ID from the filename once, before reading lines.
	base := strings.TrimSuffix(filepath.Base(file), ".jsonl")
	fileSessionID := proc.FileSessionID(base)

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	// Read the first line to extract session-level metadata (e.g. cwd for pi).
	var sessionCwd string
	if scanner.Scan() {
		firstLineCopy := make([]byte, len(scanner.Bytes()))
		copy(firstLineCopy, scanner.Bytes())
		sessionCwd = proc.SessionCwd(firstLineCopy)
	}

	// Seek back to the start so the main loop processes all lines including line 1.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek %s: %w", file, err)
	}
	scanner = bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var currentOffset int64
	for scanner.Scan() {
		currentOffset++
		if currentOffset <= lastProcessed {
			continue
		}

		stats, err := proc.ProcessEntry(scanner.Bytes(), fileSessionID, sessionCwd)
		if err != nil || stats == nil {
			continue
		}
		s.saveStats(stats)
	}

	if err := scanner.Err(); err != nil {
		slog.Warn("scanner error", "file", file, "err", err)
	}

	if currentOffset > lastProcessed {
		_, err = s.db.Exec(
			`INSERT INTO ai_sessions_watermark (tool, offset) VALUES (?, ?)
			 ON CONFLICT(tool) DO UPDATE SET offset=excluded.offset`,
			watermarkKey, currentOffset)
		if err != nil {
			return fmt.Errorf("update watermark: %w", err)
		}
	}
	return nil
}

func (s *Service) saveStats(stats *SessionStats) {
	// INSERT OR IGNORE deduplicates by the unique index on (tool, session_id, ts).
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO ai_sessions_raw
		 (tool, project, session_id, model, ts, tokens_read, tokens_write, tokens_cache, cost, tools_called, files_changed)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stats.Tool, stats.Project, stats.SessionID, stats.Model, stats.Timestamp,
		stats.TokensRead, stats.TokensWrite, stats.TokensCache, stats.Cost,
		stats.ToolsCalled, stats.FilesChanged)
	if err != nil {
		slog.Error("save stats", "err", err)
	}
}
