package summarizer

import (
	"context"
	"log/slog"
	"time"

	"github.com/mahi/flowd/internal/db"
)

// Scheduler fires BuildBlock every intervalMin minutes, aligned to clock boundaries.
type Scheduler struct {
	db          *db.DB
	intervalMin int
	// OnBlock is called after each successful block build (optional).
	OnBlock func(ctx context.Context, b *Block)
}

func NewScheduler(d *db.DB, intervalMin int) *Scheduler {
	return &Scheduler{db: d, intervalMin: intervalMin}
}

func (s *Scheduler) Run(ctx context.Context) {
	interval := time.Duration(s.intervalMin) * time.Minute

	for {
		next := nextBoundary(time.Now(), interval)
		slog.Debug("next block", "at", next.Format(time.RFC3339))

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(next)):
		}

		end := next
		start := end.Add(-interval)
		block, err := BuildBlock(ctx, s.db, start, end)
		if err != nil {
			slog.Error("build block", "err", err)
			continue
		}
		if s.OnBlock != nil {
			s.OnBlock(ctx, block)
		}
	}
}

// nextBoundary returns the next time aligned to interval from epoch.
func nextBoundary(now time.Time, interval time.Duration) time.Time {
	unix := now.UnixNano()
	iv := interval.Nanoseconds()
	remainder := unix % iv
	next := unix - remainder + iv
	return time.Unix(0, next).UTC()
}
