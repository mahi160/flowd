package fw

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WriteJournal appends a block summary to the monthly markdown log.
func WriteJournal(repoPath string, b *Block) error {
	if err := os.MkdirAll(repoPath, 0750); err != nil {
		return err
	}
	local := b.StartTS.Local()
	logFile := filepath.Join(repoPath, local.Format("2006-01")+".md")

	existing, err := os.ReadFile(logFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var sb strings.Builder
	if len(existing) == 0 {
		fmt.Fprintf(&sb, "# %s\n\n", local.Format("January 2006"))
	} else {
		sb.Write(existing)
		sb.WriteString("\n")
	}
	dayHeading := "### " + local.Format("Monday, 02 Jan")
	if !strings.Contains(string(existing), dayHeading) {
		fmt.Fprintf(&sb, "%s\n\n", dayHeading)
	}
	sb.WriteString(b.Summary)
	if b.AISummary != "" {
		sb.WriteString("\n> ")
		sb.WriteString(strings.ReplaceAll(b.AISummary, "\n", "\n> "))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	return os.WriteFile(logFile, []byte(sb.String()), 0640)
}

// PushJournal commits and pushes the journal repo with retry.
func PushJournal(ctx context.Context, repoPath, branch string) error {
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		if err := git(ctx, repoPath, "init", "-b", branch); err != nil {
			return err
		}
	}
	if err := git(ctx, repoPath, "add", "."); err != nil {
		return err
	}
	msg := "flowd: " + time.Now().UTC().Format("2006-01-02T15:04Z")
	_ = git(ctx, repoPath, "commit", "-m", msg)
	for i := range 3 {
		if err := git(ctx, repoPath, "push", "origin", branch); err == nil {
			slog.Info("journal pushed")
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(1<<i) * time.Second):
		}
	}
	return fmt.Errorf("git push failed after retries")
}

func git(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return nil
}
