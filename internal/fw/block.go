package fw

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Block struct {
	StartTS    time.Time      `json:"start"`
	EndTS      time.Time      `json:"end"`
	Project    string         `json:"project"`
	Repo       string         `json:"repo"`
	Branch     string         `json:"branch,omitempty"`
	FocusedMin int            `json:"focused_min"`
	Switches   int            `json:"switches"`
	ByTool     map[string]int `json:"by_tool"`    // minutes
	ByProject  map[string]int `json:"by_project"` // minutes
	Languages  map[string]int `json:"languages"`  // minutes (by editor time scaled to detected langs)
	FilesAdded int            `json:"files_added"`
	LinesAdded int            `json:"lines_added"`
	LinesDel   int            `json:"lines_del"`
	Summary    string         `json:"-"`
	AISummary  string         `json:"ai_summary,omitempty"`
}

// BuildBlock aggregates events in [start, end). When persist is true,
// it inserts a row into the blocks table (use false for ad-hoc previews).
func BuildBlock(ctx context.Context, d *DB, start, end time.Time, pollSec int, persist bool) (*Block, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT type, value, meta FROM events WHERE ts >= ? AND ts < ? ORDER BY ts`,
		start.UTC(), end.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type ev struct{ typ, val, meta string }
	var events []ev
	for rows.Next() {
		var e ev
		if err := rows.Scan(&e.typ, &e.val, &e.meta); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	b := &Block{
		StartTS: start, EndTS: end,
		ByTool: map[string]int{}, ByProject: map[string]int{}, Languages: map[string]int{},
	}
	repoCt := map[string]int{}
	projCt := map[string]int{}
	repoSet := map[string]struct{}{}
	editorTicks := 0
	editorByRepo := map[string]int{}
	activeTicks := 0

	secPerTick := pollSec
	if secPerTick <= 0 {
		secPerTick = 3
	}

	for _, e := range events {
		if e.typ == EvCmdChange {
			b.Switches++
		}
		if e.typ != EvActive {
			continue
		}
		activeTicks++
		var m PaneMeta
		if err := json.Unmarshal([]byte(e.meta), &m); err != nil {
			continue
		}
		b.ByTool[m.Category] += secPerTick
		if m.Repo != "" {
			repoCt[m.Repo]++
			repoSet[m.Repo+"\x00"+m.Cwd] = struct{}{}
			b.ByProject[m.Repo] += secPerTick
		}
		if m.Cwd != "" {
			projCt[m.Cwd]++
		}
		if m.Category == "editor" {
			editorTicks++
			if m.Repo != "" {
				editorByRepo[m.Repo]++
			}
		}
	}

	b.FocusedMin = (activeTicks * secPerTick) / 60
	b.Repo = topKey(repoCt)
	b.Project = topKey(projCt)

	// convert seconds → minutes (round up min 1 if any time)
	toMin := func(m map[string]int) {
		for k, s := range m {
			if s == 0 {
				delete(m, k)
				continue
			}
			m[k] = (s + 30) / 60
			if m[k] == 0 {
				m[k] = 1
			}
		}
	}
	toMin(b.ByTool)
	toMin(b.ByProject)

	// gather language + git data per unique repo
	cwdsByRepo := map[string][]string{}
	for k := range repoSet {
		parts := strings.SplitN(k, "\x00", 2)
		cwdsByRepo[parts[0]] = append(cwdsByRepo[parts[0]], parts[1])
	}
	since := start.Format("2006-01-02T15:04:05Z")
	until := end.Format("2006-01-02T15:04:05Z")
	for repo, cwds := range cwdsByRepo {
		root := RepoRoot(cwds[0])
		if root == "" {
			continue
		}
		if b.Repo == repo && b.Branch == "" {
			b.Branch = CurrentBranch(root)
		}
		stats := DiffStat(root, since, until)
		b.FilesAdded += len(stats)
		for _, s := range stats {
			b.LinesAdded += s.Added
			b.LinesDel += s.Removed
		}

		// language attribution: split editor-minutes-for-this-repo across
		// changed files, weighted by lines touched (added+removed). Falls
		// back to file count if no line data.
		repoEditorMin := editorByRepo[repo] * secPerTick / 60
		for k, v := range distributeByLines(stats, repoEditorMin) {
			b.Languages[k] += v
		}
	}

	b.Summary = render(b)

	if persist {
		dataJSON, _ := json.Marshal(b)
		if _, err := d.ExecContext(ctx,
			`INSERT INTO blocks (start_ts, end_ts, project, repo, focused_minutes, switches, data, summary)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			b.StartTS.UTC(), b.EndTS.UTC(),
			b.Project, b.Repo, b.FocusedMin, b.Switches,
			string(dataJSON), b.Summary,
		); err != nil {
			return nil, fmt.Errorf("insert block: %w", err)
		}
		slog.Info("block saved", "start", start.Format(time.RFC3339), "focused_min", b.FocusedMin)
	}
	return b, nil
}

// distributeByLines splits totalMin across languages, weighted by lines
// touched per file. If no lines are recorded (binary files only), falls
// back to even distribution by file count.
func distributeByLines(stats map[string]FileStat, totalMin int) map[string]int {
	if totalMin <= 0 || len(stats) == 0 {
		return nil
	}
	weights := map[string]int{}
	for f, s := range stats {
		w := s.Added + s.Removed
		if w == 0 {
			w = 1
		}
		weights[langOf(strings.ToLower(filepath.Ext(f)))] += w
	}
	total := 0
	for _, w := range weights {
		total += w
	}
	out := map[string]int{}
	for k, w := range weights {
		out[k] = (totalMin * w) / total
	}
	return out
}

func render(b *Block) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## %s – %s\n\n",
		b.StartTS.Local().Format("15:04"),
		b.EndTS.Local().Format("15:04"))
	fmt.Fprintf(&sb, "**Focus:** %d min  ·  **Switches:** %d\n", b.FocusedMin, b.Switches)
	if b.Repo != "" {
		if b.Branch != "" {
			fmt.Fprintf(&sb, "**Repo:** %s (%s)\n", b.Repo, b.Branch)
		} else {
			fmt.Fprintf(&sb, "**Repo:** %s\n", b.Repo)
		}
	}
	if line := topLine(b.ByProject, "min"); line != "" {
		fmt.Fprintf(&sb, "**Projects:** %s\n", line)
	}
	if line := topLine(b.ByTool, "min"); line != "" {
		fmt.Fprintf(&sb, "**Tools:** %s\n", line)
	}
	if line := topLine(b.Languages, "min"); line != "" {
		fmt.Fprintf(&sb, "**Languages:** %s\n", line)
	}
	if b.FilesAdded > 0 || b.LinesAdded > 0 || b.LinesDel > 0 {
		fmt.Fprintf(&sb, "**Code:** %d files (+%d −%d)\n", b.FilesAdded, b.LinesAdded, b.LinesDel)
	}
	return sb.String()
}

func topLine(m map[string]int, unit string) string {
	if len(m) == 0 {
		return ""
	}
	type kv struct {
		k string
		v int
	}
	var xs []kv
	for k, v := range m {
		if v == 0 {
			continue
		}
		xs = append(xs, kv{k, v})
	}
	sort.Slice(xs, func(i, j int) bool { return xs[i].v > xs[j].v })
	if len(xs) > 4 {
		xs = xs[:4]
	}
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = fmt.Sprintf("%s %d%s", x.k, x.v, unit)
	}
	return strings.Join(parts, " · ")
}

func topKey(m map[string]int) string {
	best, n := "", 0
	for k, v := range m {
		if v > n {
			best, n = k, v
		}
	}
	return best
}
