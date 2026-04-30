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

// machineDir returns the machine-specific subfolder inside the journal repo.
// Structure: <repo>/<machine>/
func machineDir(cfg *Config) string {
	return filepath.Join(expandHome(cfg.RepoPath), cfg.MachineName)
}

// WriteJournal appends a block summary to the monthly markdown log at
// <repo>/<machine>/YYYY-MM.md and regenerates the repo README.
func WriteJournal(cfg *Config, b *Block) error {
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
	if err := WriteReadme(cfg); err != nil {
		slog.Warn("readme update", "err", err)
	}
	return nil
}

// WriteReadme regenerates the README.md at the repo root with current-month stats.
func WriteReadme(cfg *Config) error {
	repoPath := expandHome(cfg.RepoPath)
	dir := machineDir(cfg)

	now := time.Now().Local()
	monthFile := filepath.Join(dir, now.Format("2006-01")+".md")

	// Count blocks this month from the monthly markdown (lightweight — no DB needed).
	// We parse the existing log file for focus lines to get stats.
	// Simpler: just show file existence + line counts as a proxy.
	// Actually, we'll use the DB for accurate stats.
	// For now: count "## " headings in the month file as block count.
	content, _ := os.ReadFile(monthFile)
	blocks := strings.Count(string(content), "\n## ")
	// count days (### headings)
	days := strings.Count(string(content), "\n### ")

	// Extract total focus minutes from block summaries.
	focusMin := 0
	for _, line := range strings.Split(string(content), "\n") {
		var f int
		if _, err := fmt.Sscanf(line, "**Focus:** %d min", &f); err == nil {
			focusMin += f
		}
	}
	focusH := focusMin / 60
	focusM := focusMin % 60

	// Top command — scan ByTool lines.
	cmdCounts := map[string]int{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimPrefix(line, "**Commands:**")
		line = strings.TrimPrefix(line, "**Tools:**")
		// lines look like: nvim 45min · claude 20min
		for _, part := range strings.Split(line, "·") {
			part = strings.TrimSpace(part)
			var name string
			var mins int
			if _, err := fmt.Sscanf(part, "%s %dmin", &name, &mins); err == nil {
				cmdCounts[name] += mins
			}
		}
	}
	topCmd := topKey(cmdCounts)

	var readme strings.Builder
	fmt.Fprintf(&readme, "# coding journal\n\n")
	fmt.Fprintf(&readme, "**%s:** %dh %dm across %d days · %d blocks\n",
		now.Format("January 2006"), focusH, focusM, days, blocks)
	if topCmd != "" {
		fmt.Fprintf(&readme, "**top command:** %s\n", topCmd)
	}
	fmt.Fprintf(&readme, "\n*Updated automatically by [flowd](https://github.com/mahi160/flowd).*\n")

	return os.WriteFile(filepath.Join(repoPath, "README.md"), []byte(readme.String()), 0640)
}

// CommitJournal stages and commits all changes in the journal repo.
// Called after every block write; the push is deferred to startup / 10 pm.
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

// PushJournal commits any pending changes and pushes to the remote with retry.
// Called on daemon startup and once per day at 10 pm.
func PushJournal(ctx context.Context, repoPath, branch string) error {
	if err := CommitJournal(ctx, repoPath, branch); err != nil {
		return err
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

func git(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return nil
}
