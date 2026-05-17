package fw

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/mahi160/flowd/internal/ai_sessions"
)

//go:embed static
var staticFiles embed.FS

type dashPayload struct {
	Period        string                    `json:"period"`
	Generated     string                    `json:"generated"`
	TotalFocusMin int                       `json:"total_focus_min"`
	TotalBlocks   int                       `json:"total_blocks"`
	TotalSwitches int                       `json:"total_switches"`
	FilesChanged  int                       `json:"files_changed"`
	LinesAdded    int                       `json:"lines_added"`
	LinesRemoved  int                       `json:"lines_removed"`
	ByProject     map[string]int            `json:"by_project"`
	ByTool        map[string]int            `json:"by_tool"`
	Languages     map[string]int            `json:"languages"`
	Heatmap       []hourBucket              `json:"heatmap"`
	Timeline      []tlBlock                 `json:"timeline"`
	StreakDays    int                       `json:"streak_days"`
	TopRepo       string                    `json:"top_repo"`
	TopBranch     string                    `json:"top_branch"`
	AIRecap       string                    `json:"ai_recap"`
	AIPerBlock    int                       `json:"ai_per_block"`
	AITools       []ai_sessions.ToolSummary `json:"ai_tools"`
	Machine       string                    `json:"machine"`
	OS            string                    `json:"os"`
}

type hourBucket struct {
	Day    string `json:"day"`
	Hour   int    `json:"hour"`
	Minute int    `json:"minute"`
}

type tlBlock struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Repo     string `json:"repo"`
	Branch   string `json:"branch"`
	Focus    int    `json:"focus"`
	Switches int    `json:"switches"`
	AI       string `json:"ai,omitempty"`
}

// aiRawRow is the DB row for ai_sessions_raw.
type aiRawRow struct {
	Tool        string
	Project     string
	SessionID   string
	Model       string
	Timestamp   time.Time
	TokensRead  int
	TokensWrite int
	TokensCache int
	Cost        float64
}

func buildDashPayload(period string, blocks []Block, aiRows []aiRawRow, streakDays int) dashPayload {
	pl := GetPlatform()
	p := dashPayload{
		Period:     period,
		Generated:  time.Now().Local().Format("Mon 02 Jan, 15:04"),
		ByProject:  map[string]int{},
		ByTool:     map[string]int{},
		Languages:  map[string]int{},
		Machine:    pl.Machine,
		OS:         pl.OS,
		StreakDays: streakDays,
	}
	repoMin := map[string]int{}
	repoBranch := map[string]string{}
	for _, b := range blocks {
		p.TotalFocusMin += b.FocusedMin
		p.TotalSwitches += b.Switches
		p.FilesChanged += b.FilesAdded
		p.LinesAdded += b.LinesAdded
		p.LinesRemoved += b.LinesDel
		for k, v := range b.ByProject {
			p.ByProject[k] += v
		}
		for k, v := range b.ByTool {
			p.ByTool[k] += v
		}
		for k, v := range b.Languages {
			p.Languages[k] += v
		}
		if b.Repo != "" {
			repoMin[b.Repo] += b.FocusedMin
			if b.Branch != "" {
				repoBranch[b.Repo] = b.Branch
			}
		}
		p.Timeline = append(p.Timeline, tlBlock{
			Start:    b.StartTS.Local().Format("15:04"),
			End:      b.EndTS.Local().Format("15:04"),
			Repo:     b.Repo,
			Branch:   b.Branch,
			Focus:    b.FocusedMin,
			Switches: b.Switches,
			AI:       b.AISummary,
		})
		if b.AISummary != "" {
			p.AIPerBlock++
		}
	}
	p.TotalBlocks = len(blocks)
	p.TopRepo = topKey(repoMin)
	p.TopBranch = repoBranch[p.TopRepo]
	p.Heatmap = buildHeatmap(period, blocks)
	p.AITools = aggregateAISessions(aiRows)
	return p
}

// aggregateAISessions groups raw rows into per-tool summaries with session detail.
func aggregateAISessions(rows []aiRawRow) []ai_sessions.ToolSummary {
	type sessKey struct {
		tool, sessionID string
	}
	type sessAccum struct {
		project              string
		model                map[string]int
		start, end           time.Time
		input, output, cache int
		cost                 float64
		messages             int
	}

	tools := map[string]*ai_sessions.ToolSummary{}
	sessions := map[sessKey]*sessAccum{}
	toolOrder := []string{}

	for _, r := range rows {
		ts, ok := tools[r.Tool]
		if !ok {
			ts = &ai_sessions.ToolSummary{
				Tool:           r.Tool,
				ModelBreakdown: map[string]int{},
			}
			tools[r.Tool] = ts
			toolOrder = append(toolOrder, r.Tool)
		}
		ts.TotalCost += r.Cost
		ts.TotalInput += r.TokensRead
		ts.TotalOutput += r.TokensWrite
		ts.TotalCache += r.TokensCache
		ts.MessageCount++
		if r.Model != "" {
			ts.ModelBreakdown[r.Model]++
		}

		sk := sessKey{r.Tool, r.SessionID}
		sa, ok := sessions[sk]
		if !ok {
			sa = &sessAccum{
				project: r.Project,
				model:   map[string]int{},
				start:   r.Timestamp,
				end:     r.Timestamp,
			}
			sessions[sk] = sa
		}
		sa.input += r.TokensRead
		sa.output += r.TokensWrite
		sa.cache += r.TokensCache
		sa.cost += r.Cost
		sa.messages++
		if r.Model != "" {
			sa.model[r.Model]++
		}
		if r.Timestamp.Before(sa.start) {
			sa.start = r.Timestamp
		}
		if r.Timestamp.After(sa.end) {
			sa.end = r.Timestamp
		}
	}

	var result []ai_sessions.ToolSummary
	for _, toolName := range toolOrder {
		ts := tools[toolName]
		ts.TopModel = topKey(ts.ModelBreakdown)
		for sk, sa := range sessions {
			if sk.tool != toolName {
				continue
			}
			ts.SessionCount++
			ts.Sessions = append(ts.Sessions, ai_sessions.AggregatedSession{
				Tool:         toolName,
				Project:      sa.project,
				SessionID:    sk.sessionID,
				Model:        topKey(sa.model),
				StartTime:    sa.start.Local().Format("15:04"),
				EndTime:      sa.end.Local().Format("15:04"),
				StartUnix:    sa.start.Unix(),
				TotalInput:   sa.input,
				TotalOutput:  sa.output,
				TotalCache:   sa.cache,
				TotalCost:    sa.cost,
				MessageCount: sa.messages,
			})
		}
		// Sort sessions newest-first using the real Unix timestamp, not the
		// formatted string (avoids midnight-crossover string sort bugs).
		sort.Slice(ts.Sessions, func(i, j int) bool {
			return ts.Sessions[i].StartUnix > ts.Sessions[j].StartUnix
		})
		result = append(result, *ts)
	}
	return result
}

func buildHeatmap(period string, blocks []Block) []hourBucket {
	if period == "week" {
		bm := map[string]map[int]int{}
		for _, b := range blocks {
			day := b.StartTS.Local().Format("Mon 02")
			h := b.StartTS.Local().Hour()
			if bm[day] == nil {
				bm[day] = map[int]int{}
			}
			bm[day][h] += b.FocusedMin
		}
		var out []hourBucket
		now := time.Now().Local()
		for i := 6; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			label := d.Format("Mon 02")
			row := bm[label]
			for h := 0; h < 24; h++ {
				out = append(out, hourBucket{Day: label, Hour: h, Minute: row[h]})
			}
		}
		return out
	}
	buckets := make([]int, 48)
	for _, b := range blocks {
		l := b.StartTS.Local()
		idx := l.Hour()*2 + l.Minute()/30
		if idx >= 0 && idx < 48 {
			buckets[idx] += b.FocusedMin
		}
	}
	out := make([]hourBucket, 48)
	for i, m := range buckets {
		out[i] = hourBucket{Day: "Today", Hour: i, Minute: m}
	}
	return out
}

// hasGitRemote returns true if a push remote is available — either from the
// config field or from an existing "origin" already set on the repo.
// Callers should cache the result for the lifetime of the daemon session.
func hasGitRemote(cfg *Config) bool {
	if cfg.GitRemote != "" {
		return true
	}
	repo := expandHome(cfg.RepoPath)
	out, err := exec.Command("git", "-C", repo, "remote", "get-url", "origin").Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

// RenderDashboard writes the static dashboard HTML.
func RenderDashboard(blocks []Block, aiRows []aiRawRow, period, aiRecap string, streakDays int, outPath string) error {
	data := buildDashPayload(period, blocks, aiRows, streakDays)
	data.AIRecap = aiRecap
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	tmpl, err := staticFiles.ReadFile("static/dashboard.html")
	if err != nil {
		return err
	}
	if !strings.Contains(string(tmpl), "__FLOWD_PAYLOAD_JSON__") {
		return fmt.Errorf("dashboard template missing __FLOWD_PAYLOAD_JSON__ placeholder")
	}
	html := strings.Replace(string(tmpl), "__FLOWD_PAYLOAD_JSON__", string(js), 1)

	if err := os.MkdirAll(filepath.Dir(outPath), 0o750); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(html), 0o644)
}

// OpenInBrowser opens a path in the user's default browser.
func OpenInBrowser(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	}
	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

// loadAIRows fetches raw AI session rows from the DB for a time range.
func loadAIRows(ctx context.Context, d *DB, start, end time.Time) ([]aiRawRow, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT tool, project, session_id, model, ts, tokens_read, tokens_write, tokens_cache, cost
		 FROM ai_sessions_raw WHERE ts >= ? AND ts < ? ORDER BY ts`, start.UTC(), end.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []aiRawRow
	for rows.Next() {
		var r aiRawRow
		if err := rows.Scan(&r.Tool, &r.Project, &r.SessionID, &r.Model, &r.Timestamp,
			&r.TokensRead, &r.TokensWrite, &r.TokensCache, &r.Cost); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
