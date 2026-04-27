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
	ByTool     map[string]int `json:"by_tool"`     // minutes
	ByProject  map[string]int `json:"by_project"`  // minutes
	Languages  map[string]int `json:"languages"`   // minutes (by editor time scaled to detected langs)
	FilesAdded int            `json:"files_added"`
	LinesAdded int            `json:"lines_added"`
	LinesDel   int            `json:"lines_del"`
	Summary    string         `json:"-"`
}

func BuildBlock(ctx context.Context, d *DB, start, end time.Time, pollSec int) (*Block, error) {
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
		a, r, f := DiffStat(root, since, until)
		b.LinesAdded += a
		b.LinesDel += r
		b.FilesAdded += f

		// language attribution: split editor minutes for this repo across changed-file extensions
		extMin := languageMinsForRepo(root, since, until, editorByRepo[repo]*secPerTick/60)
		for k, v := range extMin {
			b.Languages[k] += v
		}
	}

	b.Summary = render(b)

	dataJSON, _ := json.Marshal(b)
	_, err = d.ExecContext(ctx,
		`INSERT INTO blocks (start_ts, end_ts, project, repo, focused_minutes, switches, data, summary)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.StartTS.UTC(), b.EndTS.UTC(),
		b.Project, b.Repo, b.FocusedMin, b.Switches,
		string(dataJSON), b.Summary,
	)
	if err != nil {
		return nil, fmt.Errorf("insert block: %w", err)
	}
	slog.Info("block saved", "start", start.Format(time.RFC3339), "focused_min", b.FocusedMin)
	return b, nil
}

// languageMinsForRepo distributes totalMin minutes across the languages of files
// modified in the window, weighted by file count.
func languageMinsForRepo(root, since, until string, totalMin int) map[string]int {
	if totalMin <= 0 {
		return nil
	}
	files := changedFiles(root, since, until)
	if len(files) == 0 {
		return nil
	}
	count := map[string]int{}
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		count[langOf(ext)]++
	}
	total := 0
	for _, n := range count {
		total += n
	}
	out := map[string]int{}
	for k, n := range count {
		out[k] = (totalMin * n) / total
	}
	return out
}

func changedFiles(root, since, until string) []string {
	out, err := runOut("git", "-C", root, "log",
		"--since="+since, "--until="+until,
		"--name-only", "--pretty=format:")
	files := map[string]struct{}{}
	if err == nil {
		for _, ln := range strings.Split(out, "\n") {
			ln = strings.TrimSpace(ln)
			if ln != "" {
				files[ln] = struct{}{}
			}
		}
	}
	out, err = runOut("git", "-C", root, "diff", "--name-only")
	if err == nil {
		for _, ln := range strings.Split(out, "\n") {
			ln = strings.TrimSpace(ln)
			if ln != "" {
				files[ln] = struct{}{}
			}
		}
	}
	res := make([]string, 0, len(files))
	for f := range files {
		res = append(res, f)
	}
	return res
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
