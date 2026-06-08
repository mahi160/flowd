package fw

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// machineDir returns the machine-specific subfolder inside the journal repo.
func machineDir(cfg *Config) string {
	return filepath.Join(expandHome(cfg.RepoPath), cfg.MachineName)
}

// WriteJournal regenerates the current day's section in the monthly markdown
// log from the database (so repeated writes in the same day keep the roll-up
// accurate) and regenerates the README.
//
// File structure produced:
//
//	# January 2026
//
//	### Monday, 05 Jan
//
//	**Total:** 60 min · 2 blocks · 3 context switches
//	**Top repo:** flowd · **Languages:** Go 45min · Shell 15min
//
//	> <AI standup text, if available>
//
//	## 14:00 – 14:32
//	...per-block entry...
//
//	## 15:10 – 15:43
//	...per-block entry...
func WriteJournal(ctx context.Context, cfg *Config, d *DB, b *Block) error {
	dir := machineDir(cfg)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	local := b.StartTS.Local()
	logFile := filepath.Join(dir, local.Format("2006-01")+".md")

	// Load ALL blocks for the same calendar day so the roll-up is accurate
	// even if we're writing a mid-day block.
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	dayBlocks, err := LoadBlocks(ctx, d, dayStart, dayEnd)
	if err != nil {
		return fmt.Errorf("load day blocks: %w", err)
	}

	// Build (or re-build) the standup for today if we have AI configured.
	// We only bother when there are multiple blocks — single-block days get
	// just the raw roll-up to avoid a slow AI call per-block.
	var standupText string
	if len(dayBlocks) > 1 && cfg.AIEnabled && cfg.AICommand != "" {
		// Standup uses today + yesterday blocks; pass empty slice for yesterday
		// here since we're inside the daemon (it's a best-effort enrichment).
		s, sErr := BuildStandup(ctx, cfg, dayBlocks, nil)
		if sErr != nil {
			slog.Warn("journal standup", "err", sErr)
		} else if s != nil {
			standupText = s.Text
			if standupText == "" {
				standupText = s.Raw // fallback: show raw when AI is disabled
			}
		}
	}

	// Regenerate the entire file: keep all OTHER day sections verbatim,
	// replace (or insert) the current day's section.
	existing, err := os.ReadFile(logFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	newContent := rebuildMonthFile(string(existing), local, dayBlocks, standupText)

	if err := os.WriteFile(logFile, []byte(newContent), 0640); err != nil {
		return err
	}

	// Regenerate README — best-effort, don't fail the block write.
	if err := WriteReadme(ctx, cfg, d); err != nil {
		slog.Warn("readme update", "err", err)
	}
	return nil
}

// rebuildMonthFile takes the existing file content and replaces (or inserts)
// the section for targetDay with fresh content built from dayBlocks.
func rebuildMonthFile(existing string, targetDay time.Time, dayBlocks []Block, standupText string) string {
	dayHeading := "### " + targetDay.Format("Monday, 02 Jan")

	// Build the new day section.
	newSection := buildDaySection(targetDay, dayBlocks, standupText)

	// If the file is empty, start a new monthly file.
	if strings.TrimSpace(existing) == "" {
		var sb strings.Builder
		fmt.Fprintf(&sb, "# %s\n\n", targetDay.Format("January 2006"))
		sb.WriteString(newSection)
		return sb.String()
	}

	// Split the existing file into a map of day sections.
	sections := splitIntoSections(existing)

	// Replace (or add) our day section.
	sections[dayHeading] = newSection

	// Reconstruct: month heading first, then sections in chronological order.
	return assembleSections(existing, sections, targetDay, dayHeading)
}

// buildDaySection constructs the markdown for one calendar day.
func buildDaySection(day time.Time, blocks []Block, standupText string) string {
	if len(blocks) == 0 {
		return ""
	}
	var sb strings.Builder

	heading := "### " + day.Format("Monday, 02 Jan")
	fmt.Fprintf(&sb, "%s\n\n", heading)

	// Roll-up line.
	totalMin, totalSwitches := 0, 0
	langTotals := map[string]int{}
	repoTotals := map[string]int{}
	for _, b := range blocks {
		totalMin += b.FocusedMin
		totalSwitches += b.Switches
		for k, v := range b.Languages {
			langTotals[k] += v
		}
		if b.Repo != "" {
			repoTotals[b.Repo] += b.FocusedMin
		}
	}
	topRepo := topKey(repoTotals)

	fmt.Fprintf(&sb, "**Total:** %d min · %d blocks · %d context switches\n",
		totalMin, len(blocks), totalSwitches)
	if topRepo != "" {
		fmt.Fprintf(&sb, "**Top repo:** %s", topRepo)
		if langLine := topLine(langTotals, "min"); langLine != "" {
			fmt.Fprintf(&sb, " · **Languages:** %s", langLine)
		}
		sb.WriteString("\n")
	} else if langLine := topLine(langTotals, "min"); langLine != "" {
		fmt.Fprintf(&sb, "**Languages:** %s\n", langLine)
	}
	sb.WriteString("\n")

	// AI standup as a blockquote.
	if standupText != "" {
		sb.WriteString("> ")
		sb.WriteString(strings.ReplaceAll(strings.TrimSpace(standupText), "\n", "\n> "))
		sb.WriteString("\n\n")
	}

	// Per-block entries (sorted by start time, oldest first).
	sorted := make([]Block, len(blocks))
	copy(sorted, blocks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTS.Before(sorted[j].StartTS)
	})
	for _, b := range sorted {
		sb.WriteString(b.Summary)
		if b.AISummary != "" {
			sb.WriteString("\n> ")
			sb.WriteString(strings.ReplaceAll(b.AISummary, "\n", "\n> "))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// splitIntoSections parses a monthly markdown file into a map of
// "### Day heading" → raw section content (including the heading line).
func splitIntoSections(content string) map[string]string {
	sections := map[string]string{}
	lines := strings.Split(content, "\n")
	currentHeading := ""
	var currentLines []string

	flush := func() {
		if currentHeading != "" && len(currentLines) > 0 {
			sections[currentHeading] = strings.Join(currentLines, "\n")
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "### ") {
			flush()
			currentHeading = strings.TrimSpace(line)
			currentLines = []string{line}
		} else if currentHeading != "" {
			currentLines = append(currentLines, line)
		}
		// Lines before the first "### " heading (the "# Month" header) are
		// preserved in assembleSections by reading the original file.
	}
	flush()
	return sections
}

// assembleSections re-assembles the monthly file, keeping the month heading
// from the original content and ordering day sections chronologically.
func assembleSections(original string, sections map[string]string, targetDay time.Time, newHeading string) string {
	var sb strings.Builder

	// Preserve everything before the first "### " heading (the "# Month" line
	// and any blank lines after it).
	firstDayIdx := strings.Index(original, "\n### ")
	if firstDayIdx < 0 {
		// No existing day sections — write month header + new section.
		fmt.Fprintf(&sb, "# %s\n\n", targetDay.Format("January 2006"))
		if s, ok := sections[newHeading]; ok {
			sb.WriteString(s)
			if !strings.HasSuffix(s, "\n") {
				sb.WriteString("\n")
			}
		}
		return sb.String()
	}
	sb.WriteString(original[:firstDayIdx+1]) // includes the leading \n

	// Collect headings in the order they appeared, inserting newHeading if new.
	seen := map[string]bool{}
	var order []string
	for _, line := range strings.Split(original, "\n") {
		if strings.HasPrefix(line, "### ") {
			h := strings.TrimSpace(line)
			if !seen[h] {
				order = append(order, h)
				seen[h] = true
			}
		}
	}
	if !seen[newHeading] {
		// Insert in chronological order by parsing the date from the heading.
		inserted := false
		for i, h := range order {
			if dayFromHeading(h).After(dayFromHeading(newHeading)) {
				order = append(order[:i], append([]string{newHeading}, order[i:]...)...)
				inserted = true
				break
			}
		}
		if !inserted {
			order = append(order, newHeading)
		}
	}

	for _, heading := range order {
		if s, ok := sections[heading]; ok {
			sb.WriteString(s)
			if !strings.HasSuffix(s, "\n") {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

// dayFromHeading parses a date from a "### Monday, 02 Jan" heading.
// Returns zero time on parse failure.
func dayFromHeading(heading string) time.Time {
	// Strip "### "
	date := strings.TrimPrefix(heading, "### ")
	t, _ := time.Parse("Monday, 02 Jan", date)
	return t
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
