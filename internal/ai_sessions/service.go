package ai_sessions

import (
	"bufio"
	"context"
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
// of any per-tool failures so callers can log or surface them. Cancelling ctx
// stops the sync between files.
func (s *Service) RunSync(ctx context.Context) error {
	var errs []error
	for tool, dir := range s.paths {
		proc, ok := processors[tool]
		if !ok {
			slog.Warn("no processor for tool", "tool", tool)
			continue
		}
		if err := s.processDirectory(ctx, tool, dir, proc); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", tool, err))
		}
	}
	return errors.Join(errs...)
}

func (s *Service) processDirectory(ctx context.Context, tool, dir string, proc Processor) error {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if !d.IsDir() && filepath.Ext(path) == ".jsonl" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", dir, err)
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.processFile(tool, file, proc); err != nil {
			slog.Error("process file", "file", file, "err", err)
		}
	}
	return nil
}

// processFile processes a single .jsonl file. The watermark is keyed by
// tool + file path so each file advances independently. Files whose size
// hasn't changed since the last sync are skipped without being read —
// session transcripts are append-only, so an unchanged size means no new
// lines (this avoids re-reading the whole history every 30 minutes).
func (s *Service) processFile(tool, file string, proc Processor) error {
	watermarkKey := tool + "\x00" + file

	var lastProcessed, lastSize int64
	err := s.db.QueryRow(`SELECT offset, file_size FROM ai_sessions_watermark WHERE tool = ?`,
		watermarkKey).Scan(&lastProcessed, &lastSize)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	info, err := os.Stat(file)
	if err != nil {
		return err
	}
	if lastProcessed > 0 && info.Size() == lastSize {
		return nil // unchanged since last sync
	}

	// Extract the session ID from the filename once, before reading lines.
	base := strings.TrimSuffix(filepath.Base(file), ".jsonl")
	fileSessionID := proc.FileSessionID(base)

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// bufio.Reader + ReadBytes instead of bufio.Scanner: session lines can
	// exceed any fixed token limit (e.g. pi messages embedding images), and
	// Scanner aborts the whole file on "token too long", silently dropping
	// every line after it.
	reader := bufio.NewReaderSize(f, 64*1024)

	// One transaction per file: keeps thousands of first-import inserts fast
	// and makes the stats + watermark update atomic.
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var sessionCwd string
	var currentOffset int64
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			currentOffset++
			if currentOffset == 1 {
				// Session-level metadata (e.g. cwd for pi) lives on line 1.
				sessionCwd = proc.SessionCwd(line)
			}
			if currentOffset > lastProcessed {
				if stats, err := proc.ProcessEntry(line, fileSessionID, sessionCwd); err == nil && stats != nil {
					saveStats(tx, stats)
				}
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				slog.Warn("read session file", "file", file, "err", readErr)
			}
			break
		}
	}

	if currentOffset > lastProcessed {
		if _, err := tx.Exec(
			`INSERT INTO ai_sessions_watermark (tool, offset, file_size) VALUES (?, ?, ?)
			 ON CONFLICT(tool) DO UPDATE SET offset=excluded.offset, file_size=excluded.file_size`,
			watermarkKey, currentOffset, info.Size()); err != nil {
			return fmt.Errorf("update watermark: %w", err)
		}
	}
	return tx.Commit()
}

func saveStats(tx *sql.Tx, stats *SessionStats) {
	// INSERT OR IGNORE deduplicates by the unique index on (tool, session_id, ts).
	_, err := tx.Exec(
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
