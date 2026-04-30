package fw

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const pidFile = "/tmp/fw.pid"

func cmdStart() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the daemon (foreground)",
		RunE: func(*cobra.Command, []string) error {
			cfg, err := LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			d, err := OpenDB(cfg.DBPath())
			if err != nil {
				return err
			}
			defer d.Close()

			if data, err := os.ReadFile(pidFile); err == nil {
				if p, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
					if proc, err := os.FindProcess(p); err == nil && proc.Signal(syscall.Signal(0)) == nil {
						return fmt.Errorf("flowd already running (pid %d); run `fw stop` first", p)
					}
				}
			}
			if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
				slog.Warn("write pid", "err", err)
			}
			defer os.Remove(pidFile)

			initPlatform(cfg.MachineName)
			pl := GetPlatform()
			fmt.Println("flowd started (ctrl+c to stop, or `fw stop`)")
			slog.Info("daemon up", "db", cfg.DBPath(), "poll_sec", cfg.PollIntervalSec,
				"machine", pl.Machine, "os", pl.OS)

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				NewTracker(d, cfg.PollIntervalSec, cfg.IdleThresholdSec, cfg.WatchDirs).Run(ctx)
			}()

			go func() {
				defer wg.Done()
				runScheduler(ctx, d, cfg)
			}()

			wg.Wait()
			fmt.Println("flowd stopped")
			return nil
		},
	}
}

func cmdStop() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running daemon",
		RunE: func(*cobra.Command, []string) error {
			data, err := os.ReadFile(pidFile)
			if err != nil {
				fmt.Println("flowd not running (no pid file)")
				return nil
			}
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				return fmt.Errorf("bad pid file: %w", err)
			}
			proc, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Errorf("find process: %w", err)
			}
			if err := proc.Signal(syscall.SIGTERM); err != nil {
				fmt.Printf("could not signal %d: %v\n", pid, err)
				return nil
			}
			fmt.Printf("sent SIGTERM to flowd (pid %d)\n", pid)
			return nil
		},
	}
}

func cmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status and DB counts",
		RunE: func(*cobra.Command, []string) error {
			running := false
			pid := 0
			if data, err := os.ReadFile(pidFile); err == nil {
				if p, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
					if proc, err := os.FindProcess(p); err == nil {
						if proc.Signal(syscall.Signal(0)) == nil {
							running = true
							pid = p
						}
					}
				}
			}
			if running {
				fmt.Printf("daemon: running (pid %d)\n", pid)
			} else {
				fmt.Println("daemon: stopped")
			}

			cfg, err := LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			d, err := OpenDB(cfg.DBPath())
			if err != nil {
				return err
			}
			defer d.Close()
			var ev, bl int
			d.QueryRow("SELECT COUNT(*) FROM events").Scan(&ev)
			d.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&bl)
			fmt.Printf("db:     %s\nevents: %d\nblocks: %d\n", cfg.DBPath(), ev, bl)
			return nil
		},
	}
}

const stateBlockStart = "block_start_ts"

// runScheduler drives two independent behaviours:
//
//  1. Block cadence — a new block is written (and committed) every time the
//     user has accumulated cfg.FocusBlockMin of *focused* minutes since the
//     previous block.  blockStart is persisted in the state table so nothing
//     is lost across daemon restarts.
//
//  2. Push cadence — git push runs exactly twice per daemon lifetime:
//     • once on startup (catches any commits written during previous sessions,
//       and handles the "PC was off at 10 pm → push next morning" case), and
//     • once per calendar day at 10 pm local time.
func runScheduler(ctx context.Context, d *DB, cfg *Config) {
	// Restore blockStart from the last run, or start fresh.
	blockStart := time.Now()
	if v := d.GetState(stateBlockStart); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			// Only resume if it's within the last 24 h — older = stale.
			if time.Since(t) < 24*time.Hour {
				blockStart = t
				slog.Info("resumed block start", "from", blockStart.Format(time.RFC3339))
			}
		}
	}

	// Startup push — runs in background so it doesn't delay tracking.
	if cfg.GitRemote != "" {
		go func() {
			if err := PushJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
				slog.Warn("startup push", "err", err)
			}
		}()
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	lastPushDay := -1 // YearDay of last 10 pm push

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			// ── 10 pm push (once per calendar day, local time) ────────────
			if cfg.GitRemote != "" {
				local := now.Local()
				if local.Hour() == 22 {
					day := local.YearDay()
					if day != lastPushDay {
						lastPushDay = day
						if err := PushJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
							slog.Warn("10pm push", "err", err)
						}
					}
				}
			}

			// ── Focus-based block trigger ──────────────────────────────────
			focused, err := countFocusedMin(ctx, d, blockStart, now, cfg.PollIntervalSec)
			if err != nil {
				slog.Error("count focused min", "err", err)
				continue
			}
			if focused < cfg.FocusBlockMin {
				continue
			}

			end := now
			b, err := BuildBlock(ctx, d, blockStart, end, cfg.PollIntervalSec, true)
			if err != nil {
				slog.Error("build block", "err", err)
				continue
			}
			blockStart = end
			// Persist the new blockStart immediately so a crash doesn't orphan events.
			if err := d.SetState(stateBlockStart, blockStart.UTC().Format(time.RFC3339)); err != nil {
				slog.Warn("persist block start", "err", err)
			}

			if cfg.AIEnabled && cfg.AICommand != "" {
				ai, err := RunAI(ctx, cfg.AICommand, cfg.AIPrompt, b.Summary)
				if err != nil {
					slog.Warn("ai summary", "err", err)
				} else if ai != "" {
					b.AISummary = ai
					if _, err := d.ExecContext(ctx,
						`UPDATE blocks SET ai_summary=? WHERE start_ts=?`,
						ai, b.StartTS.UTC()); err != nil {
						slog.Warn("save ai summary", "err", err)
					}
				}
			}

			if err := WriteJournal(cfg, b); err != nil {
				slog.Error("journal write", "err", err)
				continue
			}
			if _, err := d.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
				slog.Warn("wal checkpoint", "err", err)
			}
			// Commit every block; push is deferred to startup / 10 pm.
			if cfg.GitRemote != "" {
				if err := CommitJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
					slog.Warn("journal commit", "err", err)
				}
			}
		}
	}
}

// countFocusedMin returns the number of focused minutes recorded between
// start and end by counting EvActive ticks in the DB.
func countFocusedMin(ctx context.Context, d *DB, start, end time.Time, pollSec int) (int, error) {
	var count int
	err := d.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE type=? AND ts >= ? AND ts < ?`,
		EvActive, start.UTC(), end.UTC(),
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	sec := pollSec
	if sec <= 0 {
		sec = 3
	}
	return (count * sec) / 60, nil
}
