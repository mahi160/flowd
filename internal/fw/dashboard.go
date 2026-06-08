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

// ── Types ─────────────────────────────────────────────────────────────────────

// calDay holds activity for a single calendar day (month/year heatmap).
type calDay struct {
	Date   string `json:"date"`   // "2026-05-17"
	Dow    int    `json:"dow"`    // 0=Sunday … 6=Saturday (Go time.Weekday)
	Min    int    `json:"min"`
	Blocks int    `json:"blocks"`
}

// monthBar holds aggregated activity for one calendar month (all-time chart).
type monthBar struct {
	YM    string `json:"ym"`    // "2026-05"
	Year  int    `json:"year"`
	Month int    `json:"month"` // 1-12
	Min   int    `json:"min"`
	Blocks int   `json:"blocks"`
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

// aiRawRow is a DB row for ai_sessions_raw.
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

// periodData holds all dashboard data for a single time period.
type periodData struct {
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
	CalDays       []calDay                  `json:"cal_days,omitempty"`
	MonthBars     []monthBar                `json:"month_bars,omitempty"`
	Timeline      []tlBlock                 `json:"timeline"`
	TopRepo       string                    `json:"top_repo"`
	TopBranch     string                    `json:"top_branch"`
	AITools       []ai_sessions.ToolSummary `json:"ai_tools"`
	TrackingSince string                    `json:"tracking_since,omitempty"`
	ActiveDays    int                       `json:"active_days"`
	BestDayDate   string                    `json:"best_day_date,omitempty"`
	BestDayMin    int                       `json:"best_day_min"`
}

// dashPayload is the top-level JSON blob injected into the HTML.
// It contains pre-built data for ALL periods so every tab works immediately.
type dashPayload struct {
	InitialPeriod string                 `json:"initial_period"`
	Generated     string                 `json:"generated"`
	Machine       string                 `json:"machine"`
	OS            string                 `json:"os"`
	StreakDays    int                    `json:"streak_days"`
	// Standup is the AI-generated today/yesterday standup text.
	// Empty when AI is disabled or no blocks exist for today/yesterday.
	Standup       string                 `json:"standup,omitempty"`
	// StandupRaw is the structured input (always present when there is
	// recent activity) — rendered verbatim when AI is disabled.
	StandupRaw    string                 `json:"standup_raw,omitempty"`
	// AIRecap is the legacy period-level recap (kept for backwards compat).
	AIRecap       string                 `json:"ai_recap,omitempty"`
	Periods       map[string]*periodData `json:"periods"`
}

//go:embed static
var staticFiles embed.FS

// ── Per-period builder ────────────────────────────────────────────────────────

func buildPeriodData(period string, blocks []Block, aiRows []aiRawRow, now time.Time) *periodData {
	p := &periodData{
		ByProject: map[string]int{},
		ByTool:    map[string]int{},
		Languages: map[string]int{},
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
	}
	p.TotalBlocks = len(blocks)
	p.TopRepo = topKey(repoMin)
	p.TopBranch = repoBranch[p.TopRepo]
	p.Heatmap = buildHeatmap(period, blocks)
	p.AITools = aggregateAISessions(aiRows)

	switch period {
	case "month", "year":
		p.CalDays = buildCalDays(period, blocks, now)
	case "all":
		p.MonthBars = buildMonthBars(blocks)
	}
	p.TrackingSince, p.BestDayDate, p.BestDayMin, p.ActiveDays = derivedStats(blocks)
	return p
}

// ── Top-level builder ─────────────────────────────────────────────────────────

func buildDashPayload(
	initialPeriod string,
	allBlocks map[string][]Block,
	allAIRows map[string][]aiRawRow,
	streakDays int,
) dashPayload {
	pl := GetPlatform()
	now := time.Now()
	payload := dashPayload{
		InitialPeriod: initialPeriod,
		Generated:     now.Local().Format("Mon 02 Jan, 15:04"),
		Machine:       pl.Machine,
		OS:            pl.OS,
		StreakDays:    streakDays,
		Periods:       map[string]*periodData{},
	}
	for period, blocks := range allBlocks {
		payload.Periods[period] = buildPeriodData(period, blocks, allAIRows[period], now)
	}
	return payload
}

// ── Helper builders ───────────────────────────────────────────────────────────

func buildCalDays(period string, blocks []Block, now time.Time) []calDay {
	type dd struct{ min, blocks int }
	byDate := map[string]dd{}
	for _, b := range blocks {
		key := b.StartTS.Local().Format("2006-01-02")
		d := byDate[key]
		d.min += b.FocusedMin
		d.blocks++
		byDate[key] = d
	}

	loc := now.Location()
	y, m, day := now.Date()
	var start, end time.Time
	switch period {
	case "month":
		start = time.Date(y, m, 1, 0, 0, 0, 0, loc)
		end = time.Date(y, m+1, 1, 0, 0, 0, 0, loc)
	case "year":
		start = time.Date(y, 1, 1, 0, 0, 0, 0, loc)
		end = time.Date(y+1, 1, 1, 0, 0, 0, 0, loc)
	default:
		return nil
	}
	today := time.Date(y, m, day, 0, 0, 0, 0, loc).Add(24 * time.Hour)
	if end.After(today) {
		end = today
	}

	var out []calDay
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		val := byDate[key]
		out = append(out, calDay{Date: key, Dow: int(d.Weekday()), Min: val.min, Blocks: val.blocks})
	}
	return out
}

func buildMonthBars(blocks []Block) []monthBar {
	type md struct{ min, blocks int }
	byYM := map[string]md{}
	var order []string
	for _, b := range blocks {
		ym := b.StartTS.Local().Format("2006-01")
		if _, ok := byYM[ym]; !ok {
			order = append(order, ym)
		}
		d := byYM[ym]
		d.min += b.FocusedMin
		d.blocks++
		byYM[ym] = d
	}
	sort.Strings(order)

	var out []monthBar
	for _, ym := range order {
		t, _ := time.Parse("2006-01", ym)
		d := byYM[ym]
		out = append(out, monthBar{YM: ym, Year: t.Year(), Month: int(t.Month()), Min: d.min, Blocks: d.blocks})
	}
	return out
}

func derivedStats(blocks []Block) (trackingSince, bestDayDate string, bestDayMin, activeDays int) {
	byDate := map[string]int{}
	for _, b := range blocks {
		if b.StartTS.IsZero() {
			continue
		}
		byDate[b.StartTS.Local().Format("2006-01-02")] += b.FocusedMin
	}
	for date, min := range byDate {
		if min > 0 {
			activeDays++
			if min > bestDayMin {
				bestDayMin = min
				bestDayDate = date
			}
		}
	}
	if len(blocks) > 0 {
		first := blocks[0].StartTS
		for _, b := range blocks {
			if !b.StartTS.IsZero() && b.StartTS.Before(first) {
				first = b.StartTS
			}
		}
		if !first.IsZero() {
			trackingSince = first.Local().Format("2 Jan 2006")
		}
	}
	return
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
	// today: 48 half-hour buckets
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

// aggregateAISessions groups raw rows into per-tool summaries with session detail.
func aggregateAISessions(rows []aiRawRow) []ai_sessions.ToolSummary {
	type sessKey struct{ tool, sessionID string }
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
	var toolOrder []string

	for _, r := range rows {
		ts, ok := tools[r.Tool]
		if !ok {
			ts = &ai_sessions.ToolSummary{Tool: r.Tool, ModelBreakdown: map[string]int{}}
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
			sa = &sessAccum{project: r.Project, model: map[string]int{}, start: r.Timestamp, end: r.Timestamp}
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
		sort.Slice(ts.Sessions, func(i, j int) bool {
			return ts.Sessions[i].StartUnix > ts.Sessions[j].StartUnix
		})
		result = append(result, *ts)
	}
	return result
}

// ── Rendering ─────────────────────────────────────────────────────────────────

// hasGitRemote returns true if a push remote is available.
func hasGitRemote(cfg *Config) bool {
	if cfg.GitRemote != "" {
		return true
	}
	repo := expandHome(cfg.RepoPath)
	out, err := exec.Command("git", "-C", repo, "remote", "get-url", "origin").Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

// RenderDashboard writes the single-file HTML dashboard.
func RenderDashboard(payload dashPayload, outPath string) error {
	js, err := json.Marshal(payload)
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
