package fw

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── Types ─────────────────────────────────────────────────────────────────────

// Standup is the generated standup text plus metadata.
type Standup struct {
	Text        string // AI-generated text; "" if AI is disabled
	Raw         string // structured input that was sent to AI (always populated)
	GeneratedAt time.Time
}

// ── Cache ─────────────────────────────────────────────────────────────────────

// standupCache holds the single most recent standup, keyed by an input hash,
// so repeated builds within the same daemon session don't call the AI twice.
// Only one entry is kept: keys rotate every 30 minutes, so a map would grow
// unboundedly in a long-running daemon.
var (
	standupMu       sync.Mutex
	standupCacheKey string
	standupCached   *Standup
)

func cacheKey(blocks []Block) string {
	h := sha256.New()
	// Include a 30-min time bucket so the cache self-expires through the day
	// as new commits land, without requiring an explicit invalidation.
	bucket := time.Now().Truncate(30 * time.Minute).Unix()
	fmt.Fprintf(h, "t=%d", bucket)
	for _, b := range blocks {
		// Include repo so two projects with identical minutes get distinct keys.
		fmt.Fprintf(h, "|%s:%s:%d", b.StartTS.Format("2006-01-02"), b.Repo, b.FocusedMin)
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// ── Build ─────────────────────────────────────────────────────────────────────

// BuildStandup gathers today/yesterday block data + git commits per repo,
// sends it through the configured AI command, and returns the standup.
//
// When cfg.AIEnabled is false or cfg.AICommand is empty, it still returns
// the structured raw text (useful for the journal roll-up even without AI).
//
// Results are cached by a hash of the input blocks so repeated calls within
// the same daemon session are free.
func BuildStandup(ctx context.Context, cfg *Config, todayBlocks, yestBlocks []Block) (*Standup, error) {
	allBlocks := make([]Block, 0, len(todayBlocks)+len(yestBlocks))
	allBlocks = append(allBlocks, todayBlocks...)
	allBlocks = append(allBlocks, yestBlocks...)
	if len(allBlocks) == 0 {
		return &Standup{Raw: "(no activity today or yesterday)"}, nil
	}

	key := cacheKey(allBlocks)
	standupMu.Lock()
	if key == standupCacheKey && standupCached != nil {
		cached := standupCached
		standupMu.Unlock()
		return cached, nil
	}
	standupMu.Unlock()

	raw := buildRaw(todayBlocks, yestBlocks)

	var aiText string
	if cfg.AIEnabled && cfg.AICommand != "" {
		prompt := "You are a software engineer writing your daily standup.\n" +
			"Summarise the work done today (and yesterday if shown) in 3–5 bullet points.\n" +
			"Group by project. Be concrete: mention filenames, features, or fix descriptions\n" +
			"from the commit messages. Keep it under 150 words."
		var err error
		aiText, err = RunAI(ctx, cfg.AICommand, prompt, raw)
		if err != nil {
			slog.Warn("standup AI", "err", err)
			aiText = "" // fallback: return raw only
		}
	}

	result := &Standup{
		Text:        aiText,
		Raw:         raw,
		GeneratedAt: time.Now(),
	}

	standupMu.Lock()
	standupCacheKey, standupCached = key, result
	standupMu.Unlock()

	return result, nil
}

// buildRaw constructs the structured text sent to the AI (and used as the
// journal roll-up when AI is disabled).
func buildRaw(todayBlocks, yestBlocks []Block) string {
	var sb strings.Builder

	writeDay := func(label string, blocks []Block) {
		if len(blocks) == 0 {
			return
		}
		fmt.Fprintf(&sb, "## %s\n\n", label)

		// Aggregate per project.
		type proj struct {
			focusMin     int
			filesChanged int
			linesAdded   int
			linesDel     int
		}
		byProject := map[string]*proj{}
		// Also collect the repos seen so we can fetch commits.
		repoByProject := map[string]string{}

		for _, b := range blocks {
			key := b.Repo
			if key == "" {
				key = b.Project
			}
			if key == "" {
				key = "(unknown)"
			}
			if _, ok := byProject[key]; !ok {
				byProject[key] = &proj{}
			}
			p := byProject[key]
			p.focusMin += b.FocusedMin
			p.filesChanged += b.FilesAdded
			p.linesAdded += b.LinesAdded
			p.linesDel += b.LinesDel
			if b.Repo != "" {
				repoByProject[key] = b.Repo
			}
		}

		// Determine the since time from the earliest block start.
		since := blocks[0].StartTS
		for _, b := range blocks {
			if b.StartTS.Before(since) {
				since = b.StartTS
			}
		}

		// Sort projects by focus minutes descending.
		type kv struct {
			name string
			p    *proj
		}
		var sorted []kv
		for k, v := range byProject {
			sorted = append(sorted, kv{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].p.focusMin > sorted[j].p.focusMin
		})

		for _, kv := range sorted {
			p := kv.p
			fmt.Fprintf(&sb, "### %s\n", kv.name)
			fmt.Fprintf(&sb, "- Focused: %d min", p.focusMin)
			if p.filesChanged > 0 {
				fmt.Fprintf(&sb, "  · %d files (+%d −%d lines)",
					p.filesChanged, p.linesAdded, p.linesDel)
			}
			sb.WriteString("\n")

			// Git commits for this project (best-effort).
			if repoPath := repoByProject[kv.name]; repoPath != "" {
				// Find a known cwd for this repo from the block data.
				commits := commitsForBlocks(blocks, repoPath, since)
				if len(commits) > 0 {
					sb.WriteString("- Commits:\n")
					for _, c := range commits {
						fmt.Fprintf(&sb, "  - %s: %s\n", c.Hash, c.Subject)
					}
				}
			}
			sb.WriteString("\n")
		}
	}

	writeDay("Today", todayBlocks)
	writeDay("Yesterday", yestBlocks)

	return strings.TrimSpace(sb.String())
}

// commitsForBlocks fetches recent git commits for a named repo, using the
// cwd values embedded in block events to find the repo root.
//
// b.Project holds the pane's working directory (set by topKey(projCt) in
// BuildBlock — it is a filesystem path, not a display name). We use it as
// the cwd hint for RepoRoot.
func commitsForBlocks(blocks []Block, repoName string, since time.Time) []Commit {
	for _, b := range blocks {
		if b.Repo != repoName || b.Project == "" {
			continue
		}
		root := RepoRoot(b.Project) // b.Project is a cwd path
		// b.Project is the block's top cwd overall, which may belong to a
		// different repo when a block spans several repos — verify the match
		// so commits aren't attributed to the wrong project.
		if root == "" || filepath.Base(root) != repoName {
			continue
		}
		return RecentCommits(root, since, 20)
	}
	return nil
}
