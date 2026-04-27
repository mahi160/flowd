package sync

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mahi/flowd/internal/summarizer"
)

// WriteLog appends a block summary to the monthly markdown log in repoPath.
// Files are named YYYY-MM.md and appended every 30-min block.
func WriteLog(repoPath string, block *summarizer.Block) error {
	if err := os.MkdirAll(repoPath, 0750); err != nil {
		return fmt.Errorf("create journal dir: %w", err)
	}

	local := block.StartTS.Local()
	month := local.Format("2006-01")
	logFile := filepath.Join(repoPath, month+".md")

	existing, err := os.ReadFile(logFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read log: %w", err)
	}

	var content strings.Builder
	if len(existing) == 0 {
		// Monthly header — written once
		fmt.Fprintf(&content, "# %s\n\n", local.Format("January 2006"))
	} else {
		content.Write(existing)
		content.WriteString("\n")
	}

	// Day marker when the date changes within the file
	dayHeading := "### " + local.Format("Monday, 02 Jan")
	if !strings.Contains(string(existing), dayHeading) {
		fmt.Fprintf(&content, "%s\n\n", dayHeading)
	}

	content.WriteString(block.Summary)
	content.WriteString("\n")

	if err := os.WriteFile(logFile, []byte(content.String()), 0640); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}

// Push commits and pushes the journal repo with retry.
func Push(ctx context.Context, repoPath, branch string) error {
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		if err := gitRun(ctx, repoPath, "init", "-b", branch); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
	}

	if err := gitRun(ctx, repoPath, "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	msg := fmt.Sprintf("flowd: %s", time.Now().UTC().Format("2006-01-02T15:04Z"))
	_ = gitRun(ctx, repoPath, "commit", "-m", msg) // no-op if nothing staged

	if err := gitRunWithRetry(ctx, repoPath, 3, "push", "origin", branch); err != nil {
		slog.Warn("git push failed", "err", err)
		return err
	}
	slog.Info("journal synced", "repo", repoPath)
	return nil
}

func gitRun(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return nil
}

func gitRunWithRetry(ctx context.Context, dir string, attempts int, args ...string) error {
	var lastErr error
	for i := range attempts {
		lastErr = gitRun(ctx, dir, args...)
		if lastErr == nil {
			return nil
		}
		wait := time.Duration(1<<uint(i)) * time.Second
		slog.Debug("git retry", "attempt", i+1, "wait", wait)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return lastErr
}
