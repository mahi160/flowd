package fw

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed static/dashboard.html
var dashboardHTML string

type dashPayload struct {
	Period        string         `json:"period"`
	Generated     string         `json:"generated"`
	TotalFocusMin int            `json:"total_focus_min"`
	TotalBlocks   int            `json:"total_blocks"`
	TotalSwitches int            `json:"total_switches"`
	FilesChanged  int            `json:"files_changed"`
	LinesAdded    int            `json:"lines_added"`
	LinesRemoved  int            `json:"lines_removed"`
	ByProject     map[string]int `json:"by_project"`
	ByTool        map[string]int `json:"by_tool"`
	Languages     map[string]int `json:"languages"`
	Heatmap       []hourBucket   `json:"heatmap"`     // 7×24 grid for week, 1×48 (30-min) for today
	Timeline      []tlBlock      `json:"timeline"`
	StreakDays    int            `json:"streak_days"`
	TopRepo       string         `json:"top_repo"`
	TopBranch     string         `json:"top_branch"`
	AIRecap       string         `json:"ai_recap"`
	AIPerBlock    int            `json:"ai_per_block"`
	ByMachine     map[string]int `json:"by_machine"`
	Machine       string         `json:"machine"`
	OS            string         `json:"os"`
}

type hourBucket struct {
	Day    string `json:"day"`    // "Mon 22"
	Hour   int    `json:"hour"`   // 0-23 (week) or 0-47 half-hour (today)
	Minute int    `json:"minute"` // focused minutes in the bucket
}

type tlBlock struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Repo     string `json:"repo"`
	Branch   string `json:"branch"`
	Focus    int    `json:"focus"`
	Switches int    `json:"switches"`
	Summary  string `json:"summary"`
	AI       string `json:"ai,omitempty"`
}

func buildDashPayload(period string, blocks []Block) dashPayload {
	pl := GetPlatform()
	p := dashPayload{
		Period:    period,
		Generated: time.Now().Local().Format("Mon 02 Jan, 15:04"),
		ByProject: map[string]int{},
		ByTool:    map[string]int{},
		Languages: map[string]int{},
		ByMachine: map[string]int{},
		Machine:   pl.Machine,
		OS:        pl.OS,
	}
	repoMin    := map[string]int{}
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
			Summary:  b.Summary,
			AI:       b.AISummary,
		})
		if b.AISummary != "" {
			p.AIPerBlock++
		}
		for k, v := range b.ByMachine {
			p.ByMachine[k] += v
		}
	}
	p.TotalBlocks = len(blocks)
	p.TopRepo = topKey(repoMin)
	p.TopBranch = repoBranch[p.TopRepo]
	p.Heatmap = buildHeatmap(period, blocks)
	p.StreakDays = streak(blocks)
	return p
}

func buildHeatmap(period string, blocks []Block) []hourBucket {
	if period == "week" {
		// 7 days × 24 hours
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
		// stable order: oldest day first
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

func streak(blocks []Block) int {
	if len(blocks) == 0 {
		return 0
	}
	days := map[string]bool{}
	for _, b := range blocks {
		if b.FocusedMin > 0 {
			days[b.StartTS.Local().Format("2006-01-02")] = true
		}
	}
	streak := 0
	d := time.Now().Local()
	for {
		k := d.Format("2006-01-02")
		if !days[k] {
			break
		}
		streak++
		d = d.AddDate(0, 0, -1)
	}
	return streak
}

// RenderDashboard writes a self-contained HTML file. aiRecap is optional;
// pass "" to skip the aggregate AI block in the UI.
func RenderDashboard(blocks []Block, period, aiRecap, outPath string) error {
	data := buildDashPayload(period, blocks)
	data.AIRecap = aiRecap
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	html := strings.Replace(dashboardHTML, "__FLOWD_DATA__", string(js), 1)
	if err := os.MkdirAll(filepath.Dir(outPath), 0750); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(html), 0644)
}

// OpenInBrowser opens a path/URL in the user's default browser.
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

