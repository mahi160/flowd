package fw

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// LoadBlocks fetches full Block structs (decoded from the JSON `data` column)
// for blocks whose start_ts falls in [start, end).
func LoadBlocks(ctx context.Context, d *DB, start, end time.Time) ([]Block, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT data, start_ts, end_ts, project, repo, focused_minutes, switches, summary
		 FROM blocks WHERE start_ts >= ? AND start_ts < ? ORDER BY start_ts`,
		start.UTC(), end.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Block
	for rows.Next() {
		var (
			data, p, r, sum            string
			startStr, endStr           string
			focused, switches          int
		)
		if err := rows.Scan(&data, &startStr, &endStr, &p, &r, &focused, &switches, &sum); err != nil {
			return nil, err
		}
		var b Block
		if data != "" {
			_ = json.Unmarshal([]byte(data), &b)
		}
		// fall back to flat columns if JSON missing (legacy rows)
		if b.Project == "" {
			b.Project = p
		}
		if b.Repo == "" {
			b.Repo = r
		}
		if b.FocusedMin == 0 {
			b.FocusedMin = focused
		}
		if b.Switches == 0 {
			b.Switches = switches
		}
		if b.StartTS.IsZero() {
			b.StartTS, _ = time.Parse(time.RFC3339, startStr)
		}
		if b.EndTS.IsZero() {
			b.EndTS, _ = time.Parse(time.RFC3339, endStr)
		}
		b.Summary = sum
		if b.ByTool == nil {
			b.ByTool = map[string]int{}
		}
		if b.ByProject == nil {
			b.ByProject = map[string]int{}
		}
		if b.Languages == nil {
			b.Languages = map[string]int{}
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func TextReport(blocks []Block, label string) string {
	if len(blocks) == 0 {
		return fmt.Sprintf("no blocks for %s\n", label)
	}
	var sb strings.Builder
	total := 0
	fmt.Fprintf(&sb, "=== %s ===\n\n", label)
	for _, b := range blocks {
		fmt.Fprintf(&sb, "%s → %s  %-20s focus:%dm  switches:%d\n",
			b.StartTS.Local().Format("15:04"), b.EndTS.Local().Format("15:04"),
			b.Repo, b.FocusedMin, b.Switches)
		total += b.FocusedMin
	}
	fmt.Fprintf(&sb, "\ntotal: %d blocks, %d focused minutes\n", len(blocks), total)
	return sb.String()
}

// PeriodRange returns [start, end) for "today" or "week" anchored to local time.
func PeriodRange(period string, now time.Time) (time.Time, time.Time) {
	switch period {
	case "week":
		y, m, d := now.Date()
		end := time.Date(y, m, d, 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
		return end.AddDate(0, 0, -7), end
	default:
		y, m, d := now.Date()
		start := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
		return start, start.Add(24 * time.Hour)
	}
}
