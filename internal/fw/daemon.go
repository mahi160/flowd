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
			d, err := OpenDB(cfg.DBPath)
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
			slog.Info("daemon up", "db", cfg.DBPath, "poll_sec", cfg.PollIntervalSec,
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
			d, err := OpenDB(cfg.DBPath)
			if err != nil {
				return err
			}
			defer d.Close()
			var ev, bl int
			d.QueryRow("SELECT COUNT(*) FROM events").Scan(&ev)
			d.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&bl)
			fmt.Printf("db:     %s\nevents: %d\nblocks: %d\n", cfg.DBPath, ev, bl)
			return nil
		},
	}
}

func runScheduler(ctx context.Context, d *DB, cfg *Config) {
	interval := time.Duration(cfg.SummaryIntervalMin) * time.Minute
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
		b, err := BuildBlock(ctx, d, start, end, cfg.PollIntervalSec, true)
		if err != nil {
			slog.Error("build block", "err", err)
			continue
		}
		if b.FocusedMin < cfg.MinFocusMin {
			slog.Info("block below threshold, skipping push",
				"focused_min", b.FocusedMin, "min", cfg.MinFocusMin)
			continue
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
		if err := WriteJournal(cfg.RepoPath, b); err != nil {
			slog.Error("journal write", "err", err)
			continue
		}
		if _, err := d.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			slog.Warn("wal checkpoint", "err", err)
		}
		if cfg.GitRemote != "" {
			if err := PushJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
				slog.Warn("journal push", "err", err)
			}
		}
	}
}

func nextBoundary(now time.Time, interval time.Duration) time.Time {
	u := now.UnixNano()
	iv := interval.Nanoseconds()
	return time.Unix(0, u-(u%iv)+iv).UTC()
}
