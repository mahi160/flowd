package summarizer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mahi/flowd/internal/db"
	"github.com/mahi/flowd/internal/session"
)

type Block struct {
	StartTS        time.Time
	EndTS          time.Time
	Project        string
	Repo           string
	FocusedMinutes int
	KeyCount       int
	Switches       int
	Tools          []string
	Summary        string
}

// BuildBlock aggregates events in [start, end) into a Block and persists it.
func BuildBlock(ctx context.Context, d *db.DB, start, end time.Time) (*Block, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT type, value, meta FROM events WHERE ts >= ? AND ts < ? ORDER BY ts`,
		start.UTC(), end.UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	type row struct {
		typ, value, meta string
	}
	var events []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.typ, &r.value, &r.meta); err != nil {
			return nil, err
		}
		events = append(events, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	block := &Block{StartTS: start, EndTS: end}
	toolSet := map[string]struct{}{}
	repoCount := map[string]int{}
	projectCount := map[string]int{}
	activeTicks := 0

	for _, e := range events {
		if e.typ == string(session.EventPaneActive) {
			activeTicks++
			var meta session.PaneMeta
			if err := json.Unmarshal([]byte(e.meta), &meta); err == nil {
				if meta.Category != "" {
					toolSet[meta.Category] = struct{}{}
				}
				if meta.Repo != "" {
					repoCount[meta.Repo]++
				}
				if meta.Cwd != "" {
					projectCount[meta.Cwd]++
				}
			}
		}
		if e.typ == string(session.EventCmdChange) {
			block.Switches++
		}
	}

	// poll ticks × poll_interval ≈ focused seconds; assume 3s default
	block.FocusedMinutes = (activeTicks * 3) / 60

	block.Repo = topKey(repoCount)
	block.Project = topKey(projectCount)

	for t := range toolSet {
		block.Tools = append(block.Tools, t)
	}

	block.Summary = templateSummary(block)

	toolsJSON, _ := json.Marshal(block.Tools)
	_, err = d.ExecContext(ctx,
		`INSERT INTO blocks (start_ts, end_ts, project, repo, focused_minutes, key_count, switches, tools, summary)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		block.StartTS.UTC(), block.EndTS.UTC(),
		block.Project, block.Repo,
		block.FocusedMinutes, block.KeyCount, block.Switches,
		string(toolsJSON), block.Summary,
	)
	if err != nil {
		return nil, fmt.Errorf("insert block: %w", err)
	}

	slog.Info("block saved", "start", start.Format(time.RFC3339), "focused_min", block.FocusedMinutes)
	return block, nil
}

func templateSummary(b *Block) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## %s – %s\n\n",
		b.StartTS.Format("15:04"),
		b.EndTS.Format("15:04"),
	)
	if b.Repo != "" {
		fmt.Fprintf(&sb, "**Repo:** %s\n", b.Repo)
	}
	fmt.Fprintf(&sb, "**Focus:** %d min\n", b.FocusedMinutes)
	fmt.Fprintf(&sb, "**Switches:** %d\n", b.Switches)
	if len(b.Tools) > 0 {
		fmt.Fprintf(&sb, "**Tools:** %s\n", strings.Join(b.Tools, ", "))
	}
	return sb.String()
}

func topKey(m map[string]int) string {
	best, bestN := "", 0
	for k, n := range m {
		if n > bestN {
			best, bestN = k, n
		}
	}
	return best
}
