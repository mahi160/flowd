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
	Machine    string         `json:"machine,omitempty"`
	OS         string         `json:"os,omitempty"`
	ByMachine  map[string]int `json:"by_machine,omitempty"` // minutes per machine (multi-device)
	Summary    string         `json:"-"`
	AISummary  string         `json:"ai_summary,omitempty"`
}

// BuildBlock aggregates events in [start, end). When persist is true,
// it inserts a row into the blocks table (use false for ad-hoc previews).
func BuildBlock(ctx context.Context, d *DB, start, end time.Time, pollSec int, persist bool) (*Block, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT type, meta FROM events WHERE ts >= ? AND ts < ? ORDER BY ts`,
		start.UTC(), end.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	b := &Block{
		StartTS: start, EndTS: end,
		ByTool: map[string]int{}, ByProject: map[string]int{},
		Languages: map[string]int{}, ByMachine: map[string]int{},
	}
	repoCt := map[string]int{}
	projCt := map[string]int{}
	repoCwd := map[string]string{}    // repo → a cwd inside it (for RepoRoot lookup)
	editorByRepo := map[string]int{}  // repo → editor ticks (git-diff path)
	runtimeByCmd := map[string]int{}  // command → runtime ticks (for lang inference)
	cwdNoRepo := map[string]int{}     // cwd → editor ticks (no git repo)
	nvimLangTicks := map[string]int{} // filetype → ticks (from nvim plugin; bypasses git-diff)
	activeTicks := 0

	secPerTick := pollSec
	if secPerTick <= 0 {
		secPerTick = 3
	}

	for rows.Next() {
		var typ, meta string
		if err := rows.Scan(&typ, &meta); err != nil {
			return nil, err
		}
		if typ == EvSessionChange {
			b.Switches++
		}
		if typ != EvActive {
			continue
		}
		activeTicks++
		var m PaneMeta
		if err := json.Unmarshal([]byte(meta), &m); err != nil {
			slog.Warn("unmarshal event meta", "err", err)
			continue
		}
		cmd := m.Command
		if cmd == "" {
			cmd = m.Category // fallback for legacy events
		}
		b.ByTool[cmd] += secPerTick
		if m.Machine != "" {
			b.ByMachine[m.Machine] += secPerTick
		}
		if m.Repo != "" {
			repoCt[m.Repo]++
			repoCwd[m.Repo] = m.Cwd
			b.ByProject[m.Repo] += secPerTick
		}
		if m.Cwd != "" {
			projCt[m.Cwd]++
		}
		if m.Category == "editor" {
			if m.NvimFiletype != "" {
				// Plugin provided an exact filetype — collect separately so we
				// don't also run git-diff attribution for the same ticks.
				nvimLangTicks[m.NvimFiletype]++
			} else if m.Repo != "" {
				editorByRepo[m.Repo]++
			} else if m.Cwd != "" {
				cwdNoRepo[m.Cwd]++
			}
		}
		if m.Category == "runtime" {
			runtimeByCmd[m.Command]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	b.FocusedMin = (activeTicks * secPerTick) / 60
	b.Repo = topKey(repoCt)
	b.Project = topKey(projCt)

	// convert seconds → minutes using truncating division (same as FocusedMin).
	// Drop entries under 30 seconds rather than inflating them to 1 min — that
	// would cause the per-tool/project sums to exceed FocusedMin.
	toMin := func(m map[string]int) {
		for k, s := range m {
			m[k] = s / 60
			if m[k] == 0 {
				delete(m, k)
			}
		}
	}
	toMin(b.ByTool)
	toMin(b.ByProject)
	toMin(b.ByMachine)
	b.Machine = topKey(b.ByMachine)
	// OS from current platform (events may mix machines; use daemon's own OS)
	if pl := GetPlatform(); pl != nil {
		b.OS = pl.OS
	}

	// gather language + git data per unique repo
	// Always use UTC when passing timestamps to git --since/--until.
	// The format string has a literal 'Z' suffix, so the time MUST be UTC first.
	since := start.UTC().Format("2006-01-02T15:04:05Z")
	until := end.UTC().Format("2006-01-02T15:04:05Z")
	for repo, cwd := range repoCwd {
		root := RepoRoot(cwd)
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

	// nvim plugin filetype → language minutes (highest fidelity; no git needed).
	for ft, ticks := range nvimLangTicks {
		if lang := langFromFiletype(ft); lang != "" {
			b.Languages[lang] += (ticks * secPerTick) / 60
		}
	}

	// runtime command → language (node → JS, python → Python, etc.)
	for cmd, ticks := range runtimeByCmd {
		if lang := LangFromCommand(cmd); lang != "" {
			b.Languages[lang] += (ticks * secPerTick) / 60 // truncating, consistent with toMin
		}
	}

	// editor in dirs with no git repo → scan cwd for dominant language
	for cwd, ticks := range cwdNoRepo {
		mins := (ticks * secPerTick) / 60
		if mins == 0 {
			continue // drop sub-minute editor sessions
		}
		counts := ScanLangs(cwd)
		total := 0
		for _, n := range counts {
			total += n
		}
		if total > 0 {
			for lang, n := range counts {
				b.Languages[lang] += (mins * n) / total
			}
		}
	}

	// Clean up Languages: remove noise extensions and zero-value entries.
	// Known-good languages are in extLang; anything else is dropped.
	for lang, min := range b.Languages {
		if min <= 0 || !isKnownLang(lang) {
			delete(b.Languages, lang)
		}
	}

	b.Summary = render(b)

	if persist {
		dataJSON, err := json.Marshal(b)
		if err != nil {
			slog.Error("marshal block", "err", err)
			return nil, fmt.Errorf("marshal block: %w", err)
		}
		if _, err := d.ExecContext(ctx,
			`INSERT INTO blocks (start_ts, end_ts, project, repo, focused_minutes, switches, data, summary)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(start_ts) DO UPDATE SET
			   end_ts=excluded.end_ts, project=excluded.project, repo=excluded.repo,
			   focused_minutes=excluded.focused_minutes, switches=excluded.switches,
			   data=excluded.data, summary=excluded.summary
			   -- ai_summary deliberately excluded: preserve any AI summary already written`,
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
	fmt.Fprintf(&sb, "**Focus:** %d min  ·  **Context switches:** %d\n", b.FocusedMin, b.Switches)
	if b.Machine != "" {
		fmt.Fprintf(&sb, "**Machine:** %s (%s)\n", b.Machine, b.OS)
	}
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
