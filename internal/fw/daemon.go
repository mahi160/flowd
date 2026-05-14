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

	"github.com/mahi160/flowd/internal/ai_sessions"
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

			go func() {
				defer wg.Done()
				// Run periodic AI scan every 30 mins
				ticker := time.NewTicker(30 * time.Minute)
				defer ticker.Stop()
				svc := ai_sessions.NewService(d.DB, cfg.AISessionPaths)
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if err := svc.RunSync(); err != nil {
							slog.Error("ai session scan", "err", err)
						}
					}
				}
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

// runScheduler drives two independent cadences:
//
//  1. Commit cadence — every 30 minutes a new block is built from accumulated
//     events and committed to git (if a remote is configured).
//
//  2. Push cadence — git push runs once per hour, independently of commits.
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

	commitTicker := time.NewTicker(30 * time.Minute)
	defer commitTicker.Stop()

	pushTicker := time.NewTicker(60 * time.Minute)
	defer pushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case now := <-commitTicker.C:
			// ── Build block + commit every 30 minutes ─────────────────────
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
			if cfg.GitRemote != "" {
				if err := CommitJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
					slog.Warn("journal commit", "err", err)
				}
			}

		case <-pushTicker.C:
			// ── Push every hour ───────────────────────────────────────────
			if cfg.GitRemote != "" {
				if err := PushJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
					slog.Warn("hourly push", "err", err)
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
