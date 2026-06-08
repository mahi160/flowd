package fw

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// machineDir returns the machine-specific subfolder inside the journal repo.
func machineDir(cfg *Config) string {
	return filepath.Join(expandHome(cfg.RepoPath), cfg.MachineName)
}

// WriteJournal appends a block summary to the monthly markdown log and
// regenerates the repo README using live DB stats.
func WriteJournal(ctx context.Context, cfg *Config, d *DB, b *Block) error {
	dir := machineDir(cfg)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	local := b.StartTS.Local()
	logFile := filepath.Join(dir, local.Format("2006-01")+".md")

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
	if err := os.WriteFile(logFile, []byte(sb.String()), 0640); err != nil {
		return err
	}

	// Regenerate README — best-effort, don't fail the block write.
	if err := WriteReadme(ctx, cfg, d); err != nil {
		slog.Warn("readme update", "err", err)
	}
	return nil
}

// WriteReadme regenerates the README.md at the repo root using current-month
// stats queried directly from the DB (avoids fragile markdown parsing).
func WriteReadme(ctx context.Context, cfg *Config, d *DB) error {
	repoPath := expandHome(cfg.RepoPath)
	totalBlocks, totalDays, focusMin := d.QueryMonthStats(ctx)

	now := time.Now()
	focusH := focusMin / 60
	focusM := focusMin % 60

	var readme strings.Builder
	fmt.Fprintf(&readme, "# coding journal\n\n")
	fmt.Fprintf(&readme, "**%s:** %dh %dm across %d days · %d blocks\n",
		now.Format("January 2006"), focusH, focusM, totalDays, totalBlocks)
	fmt.Fprintf(&readme, "\n*Updated automatically by [flowd](https://github.com/mahi160/flowd).*\n")

	return os.WriteFile(filepath.Join(repoPath, "README.md"), []byte(readme.String()), 0640)
}

// CommitJournal stages and commits all changes in the journal repo.
func CommitJournal(ctx context.Context, repoPath, branch string) error {
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		if err := git(ctx, repoPath, "init", "-b", branch); err != nil {
			return err
		}
	}
	if err := git(ctx, repoPath, "add", "."); err != nil {
		return err
	}
	msg := "flowd: " + time.Now().UTC().Format("2006-01-02T15:04Z")
	_ = git(ctx, repoPath, "commit", "-m", msg) // no-op if nothing to commit
	return nil
}

// PushJournal commits any pending changes, rebases on remote, then pushes.
func PushJournal(ctx context.Context, repoPath, branch string) error {
	if err := CommitJournal(ctx, repoPath, branch); err != nil {
		return err
	}
	// Rebase on remote before pushing to avoid non-fast-forward rejections.
	if out, err := runGit(ctx, repoPath, "pull", "--rebase", "origin", branch); err != nil {
		slog.Warn("journal pull --rebase", "out", strings.TrimSpace(out))
		// Not fatal — try pushing anyway (e.g. no remote commits yet).
	}
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
