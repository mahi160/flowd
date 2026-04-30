package fw

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type BlockRow struct {
	StartTS, EndTS string
	Repo           string
	FocusedMin     int
	Switches       int
	Summary        string
}

func QueryBlocks(ctx context.Context, d *DB, start, end time.Time) ([]BlockRow, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT start_ts, end_ts, repo, focused_minutes, switches, summary
		 FROM blocks WHERE start_ts >= ? AND start_ts < ? ORDER BY start_ts`,
		start.UTC(), end.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BlockRow
	for rows.Next() {
		var b BlockRow
		if err := rows.Scan(&b.StartTS, &b.EndTS, &b.Repo, &b.FocusedMin, &b.Switches, &b.Summary); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func TextReport(blocks []BlockRow, label string) string {
	if len(blocks) == 0 {
		return fmt.Sprintf("no blocks for %s\n", label)
	}
	var sb strings.Builder
	total := 0
	fmt.Fprintf(&sb, "=== %s ===\n\n", label)
	for _, b := range blocks {
		fmt.Fprintf(&sb, "%s → %s  %-20s focus:%dm  switches:%d\n",
			b.StartTS, b.EndTS, b.Repo, b.FocusedMin, b.Switches)
		total += b.FocusedMin
	}
	fmt.Fprintf(&sb, "\ntotal: %d blocks, %d focused minutes\n", len(blocks), total)
	return sb.String()
}
