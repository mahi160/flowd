package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mahi/flowd/internal/db"
)

type BlockRow struct {
	StartTS        string
	EndTS          string
	Repo           string
	FocusedMinutes int
	Switches       int
	Tools          string
	Summary        string
}

func QueryBlocks(ctx context.Context, d *db.DB, start, end time.Time) ([]BlockRow, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT start_ts, end_ts, repo, focused_minutes, switches, tools, summary
		 FROM blocks WHERE start_ts >= ? AND start_ts < ? ORDER BY start_ts`,
		start.UTC(), end.UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []BlockRow
	for rows.Next() {
		var b BlockRow
		if err := rows.Scan(&b.StartTS, &b.EndTS, &b.Repo, &b.FocusedMinutes, &b.Switches, &b.Tools, &b.Summary); err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

func TextReport(blocks []BlockRow, label string) string {
	if len(blocks) == 0 {
		return fmt.Sprintf("no blocks for %s\n", label)
	}
	var sb strings.Builder
	totalFocus := 0
	fmt.Fprintf(&sb, "=== %s ===\n\n", label)
	for _, b := range blocks {
		fmt.Fprintf(&sb, "%s → %s  %-20s focus:%dm  switches:%d\n",
			b.StartTS, b.EndTS, b.Repo, b.FocusedMinutes, b.Switches)
		totalFocus += b.FocusedMinutes
	}
	fmt.Fprintf(&sb, "\ntotal: %d blocks, %d focused minutes\n", len(blocks), totalFocus)
	return sb.String()
}

func HTMLReport(blocks []BlockRow, label string) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8">`)
	fmt.Fprintf(&sb, `<title>Flowd — %s</title>`, label)
	sb.WriteString(`<style>
body{font-family:monospace;max-width:900px;margin:2rem auto;padding:0 1rem;background:#0d1117;color:#c9d1d9}
h1{color:#58a6ff}table{width:100%;border-collapse:collapse}
th,td{text-align:left;padding:.4rem .8rem;border-bottom:1px solid #21262d}
th{background:#161b22;color:#8b949e}tr:hover td{background:#161b22}
.focus{color:#3fb950}.repo{color:#d2a8ff}
</style></head><body>`)
	fmt.Fprintf(&sb, "<h1>Flowd — %s</h1>\n", label)

	if len(blocks) == 0 {
		sb.WriteString("<p>No blocks recorded.</p>")
	} else {
		totalFocus := 0
		sb.WriteString(`<table><tr><th>Start</th><th>End</th><th>Repo</th><th>Focus (min)</th><th>Switches</th><th>Tools</th></tr>`)
		for _, b := range blocks {
			fmt.Fprintf(&sb,
				`<tr><td>%s</td><td>%s</td><td class="repo">%s</td><td class="focus">%d</td><td>%d</td><td>%s</td></tr>`,
				b.StartTS, b.EndTS, b.Repo, b.FocusedMinutes, b.Switches, b.Tools,
			)
			totalFocus += b.FocusedMinutes
		}
		sb.WriteString("</table>")
		fmt.Fprintf(&sb, "<p><strong>Total:</strong> %d blocks, <span class=\"focus\">%d focused minutes</span></p>", len(blocks), totalFocus)
	}

	sb.WriteString("</body></html>")
	return sb.String()
}
