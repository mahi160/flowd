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

// WriteLog appends a block summary to a daily markdown log in repoPath.
func WriteLog(repoPath string, block *summarizer.Block) error {
	if err := os.MkdirAll(repoPath, 0750); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}

	date := block.StartTS.Local().Format("2006-01-02")
	logFile := filepath.Join(repoPath, date+".md")

	// Ensure daily header exists
	existing, err := os.ReadFile(logFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read log: %w", err)
	}

	var content strings.Builder
	if len(existing) == 0 {
		fmt.Fprintf(&content, "# %s\n\n", date)
	} else {
		content.Write(existing)
		content.WriteString("\n")
	}
	content.WriteString(block.Summary)
	content.WriteString("\n")

	if err := os.WriteFile(logFile, []byte(content.String()), 0640); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}

// Push commits and pushes the repo with retry.
func Push(ctx context.Context, repoPath, branch string) error {
	// init repo if needed
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		if err := gitRun(ctx, repoPath, "init", "-b", branch); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
	}

	if err := gitRun(ctx, repoPath, "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	msg := fmt.Sprintf("flowd: %s", time.Now().UTC().Format("2006-01-02T15:04Z"))
	// commit may fail if nothing staged — that's fine
	_ = gitRun(ctx, repoPath, "commit", "-m", msg)

	if err := gitRunWithRetry(ctx, repoPath, 3, "push", "origin", branch); err != nil {
		slog.Warn("git push failed", "err", err)
		return err
	}
	slog.Info("synced", "repo", repoPath)
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
