package fw

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── Types ─────────────────────────────────────────────────────────────────────

// StandupProject is the per-project summary fed to the AI.
type StandupProject struct {
	Name       string
	FocusedMin int
	FilesChanged int
	LinesAdded  int
	LinesDel    int
	Commits     []Commit
}

// Standup is the generated standup text plus metadata.
type Standup struct {
	Text      string    // AI-generated text; "" if AI is disabled
	Raw       string    // structured input that was sent to AI (always populated)
	GeneratedAt time.Time
}

// ── Cache ─────────────────────────────────────────────────────────────────────

// standupCache caches the generated standup keyed by a day-hash so multiple
// `fw dashboard` invocations on the same day don't call the AI twice.
var (
	standupMu    sync.Mutex
	standupCache = map[string]*Standup{}
)

func cacheKey(blocks []Block) string {
	h := sha256.New()
	for _, b := range blocks {
		fmt.Fprintf(h, "%s%d", b.StartTS.Format("2006-01-02"), b.FocusedMin)
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
	allBlocks := append(todayBlocks, yestBlocks...)
	if len(allBlocks) == 0 {
		return &Standup{Raw: "(no activity today or yesterday)"}, nil
	}

	key := cacheKey(allBlocks)
	standupMu.Lock()
	if cached, ok := standupCache[key]; ok {
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
	standupCache[key] = result
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
func commitsForBlocks(blocks []Block, repoName string, since time.Time) []Commit {
	// Find any block for this repo and extract a cwd we can use as a root hint.
	for _, b := range blocks {
		root := ""
		if b.Repo == repoName && b.Project != "" {
			root = RepoRoot(b.Project)
		}
		if root == "" {
			continue
		}
		return RecentCommits(root, since, 20)
	}
	return nil
}
