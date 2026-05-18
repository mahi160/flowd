package fw

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mahi160/flowd/internal/ai_sessions"
	"github.com/spf13/cobra"
)

const pidFile = "/tmp/fw.pid"

func defaultLogPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "flowd", "flowd.log")
}

func cmdStart() *cobra.Command {
	var daemonMode bool
	c := &cobra.Command{
		Use:   "start",
		Short: "Start the tracking daemon in the background",
		RunE: func(*cobra.Command, []string) error {
			if daemonMode {
				return runDaemon()
			}
			return spawnDaemon()
		},
	}
	c.Flags().BoolVar(&daemonMode, "daemon", false, "")
	_ = c.Flags().MarkHidden("daemon")
	return c
}

// spawnDaemon re-execs fw with --daemon, detached from the terminal.
func spawnDaemon() error {
	if data, err := os.ReadFile(pidFile); err == nil {
		if p, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			if proc, err := os.FindProcess(p); err == nil && proc.Signal(syscall.Signal(0)) == nil {
				return fmt.Errorf("flowd already running (pid %d); run `fw stop` first", p)
			}
		}
	}

	logPath := defaultLogPath()
	if err := os.MkdirAll(filepath.Dir(logPath), 0750); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	args := append(os.Args[1:], "--daemon")
	cmd := exec.Command(exe, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("spawn daemon: %w", err)
	}

	fmt.Printf("flowd started (pid %d)\n", cmd.Process.Pid)
	fmt.Printf("logs: %s\n", logPath)
	return nil
}

// runDaemon is the actual daemon loop — only called by the spawned child.
func runDaemon() error {
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return err
	}
	d, err := OpenDB(cfg.DBPath())
	if err != nil {
		return err
	}
	defer d.Close()

	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		slog.Warn("write pid", "err", err)
	}
	defer os.Remove(pidFile)

	initPlatform(cfg.MachineName)
	pl := GetPlatform()
	slog.Info("daemon up", "db", cfg.DBPath(), "poll_sec", cfg.PollIntervalSec,
		"machine", pl.Machine, "os", pl.OS)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)

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
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		svc := ai_sessions.NewService(d.DB, cfg.AISessionPaths)
		if err := svc.RunSync(); err != nil {
			slog.Warn("initial ai session scan", "err", err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := svc.RunSync(); err != nil {
					slog.Warn("ai session scan", "err", err)
				}
			}
		}
	}()

	wg.Wait()
	slog.Info("daemon stopped")
	return nil
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

// runScheduler drives three independent cadences:
//
//  1. Focus cadence  — every FocusBlockMin of focused time triggers a block
//     build + git commit. Wall-clock idle time does NOT count.
//
//  2. Push cadence   — git push runs once per hour.
//
//  3. Cleanup cadence — raw events older than 90 days are pruned daily.
func runScheduler(ctx context.Context, d *DB, cfg *Config) {
	// Resolve the remote once — it won't change while the daemon runs.
	hasRemote := hasGitRemote(cfg)

	blockStart := time.Now()
	if v := d.GetState(stateBlockStart); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			if time.Since(t) < 24*time.Hour {
				blockStart = t
				slog.Info("resumed block start", "from", blockStart.Format(time.RFC3339))
			}
		}
	}

	focusThreshold := cfg.FocusBlockMin
	if focusThreshold <= 0 {
		focusThreshold = 30
	}

	focusTicker := time.NewTicker(1 * time.Minute)
	defer focusTicker.Stop()

	pushTicker := time.NewTicker(60 * time.Minute)
	defer pushTicker.Stop()

	// Prune old events once at startup, then daily.
	pruneEvents := func() {
		if err := d.PruneEvents(ctx, 90); err != nil {
			slog.Warn("prune events", "err", err)
		}
	}
	pruneEvents()
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case now := <-focusTicker.C:
			focused, err := countFocusedMin(ctx, d, blockStart, now, cfg.PollIntervalSec)
			if err != nil {
				slog.Warn("count focused min", "err", err)
				continue
			}
			if focused < focusThreshold {
				continue
			}

			slog.Info("focus threshold reached", "focused_min", focused, "threshold", focusThreshold)
			end := now
			b, err := BuildBlock(ctx, d, blockStart, end, cfg.PollIntervalSec, true)
			if err != nil {
				slog.Error("build block", "err", err)
				continue
			}
			blockStart = end
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

			if err := WriteJournal(ctx, cfg, d, b); err != nil {
				slog.Error("journal write", "err", err)
				continue
			}
			if _, err := d.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
				slog.Warn("wal checkpoint", "err", err)
			}
			if hasRemote {
				if err := CommitJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
					slog.Warn("journal commit", "err", err)
				}
			}

		case <-pushTicker.C:
			if hasRemote {
				if err := PushJournal(ctx, cfg.RepoPath, cfg.Branch); err != nil {
					slog.Warn("hourly push", "err", err)
				}
			}

		case <-cleanupTicker.C:
			pruneEvents()
		}
	}
}

// countFocusedMin counts EvActive ticks between start and end.
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
