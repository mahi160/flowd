package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/mahi/flowd/internal/config"
	"github.com/mahi/flowd/internal/db"
	"github.com/mahi/flowd/internal/logger"
	"github.com/mahi/flowd/internal/session"
	"github.com/mahi/flowd/internal/summarizer"
	flowsync "github.com/mahi/flowd/internal/sync"
)

var (
	cfgPath string
	debug   bool
)

func main() {
	root := &cobra.Command{
		Use:   "fw",
		Short: "Flowd - local coding memory daemon",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.Init(debug)
		},
	}

	root.PersistentFlags().StringVar(&cfgPath, "config", config.DefaultPath(), "config file path")
	root.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	root.AddCommand(
		cmdStart(),
		cmdStop(),
		cmdStatus(),
		cmdSummary(),
		cmdReport(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadConfig() (*config.Config, error) {
	return config.Load(cfgPath)
}

func openDB(cfg *config.Config) (*db.DB, error) {
	return db.Open(cfg.DBPath)
}

func cmdStart() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the Flowd daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			d, err := openDB(cfg)
			if err != nil {
				return err
			}
			defer d.Close()

			fmt.Println("flowd daemon starting (foreground mode, ctrl+c to stop)")
			logger.L.Info("daemon started", "db", cfg.DBPath, "poll_sec", cfg.PollIntervalSec)

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				tracker := session.NewTracker(d, cfg.PollIntervalSec)
				tracker.Run(ctx)
			}()

			go func() {
				defer wg.Done()
				sched := summarizer.NewScheduler(d, cfg.SummaryIntervalMin)
				sched.OnBlock = func(ctx context.Context, b *summarizer.Block) {
					if err := flowsync.WriteLog(cfg.RepoPath, b); err != nil {
						logger.L.Error("write log", "err", err)
						return
					}
					if cfg.PushDB {
						if err := flowsync.Push(ctx, cfg.RepoPath, cfg.Branch); err != nil {
							logger.L.Warn("sync push failed", "err", err)
						}
					}
				}
				sched.Run(ctx)
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
		Short: "Stop the Flowd daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("stop not yet implemented")
			return nil
		},
	}
}

func cmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon and DB status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			d, err := openDB(cfg)
			if err != nil {
				return err
			}
			defer d.Close()

			var evCount, blockCount int
			d.QueryRow("SELECT COUNT(*) FROM events").Scan(&evCount)
			d.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&blockCount)

			fmt.Printf("db:     %s\n", cfg.DBPath)
			fmt.Printf("events: %d\n", evCount)
			fmt.Printf("blocks: %d\n", blockCount)
			return nil
		},
	}
}

func cmdSummary() *cobra.Command {
	return &cobra.Command{
		Use:   "summary [now]",
		Short: "Generate summary for the current or last 30-min block",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			d, err := openDB(cfg)
			if err != nil {
				return err
			}
			defer d.Close()

			now := time.Now().UTC()
			interval := time.Duration(cfg.SummaryIntervalMin) * time.Minute
			// align end to last boundary
			unix := now.UnixNano()
			iv := interval.Nanoseconds()
			endNano := unix - (unix % iv)
			end := time.Unix(0, endNano).UTC()
			start := end.Add(-interval)

			block, err := summarizer.BuildBlock(cmd.Context(), d, start, end)
			if err != nil {
				return err
			}
			fmt.Println(block.Summary)
			return nil
		},
	}
}

func cmdReport() *cobra.Command {
	return &cobra.Command{
		Use:   "report [today|week]",
		Short: "Show activity report",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			d, err := openDB(cfg)
			if err != nil {
				return err
			}
			defer d.Close()

			period := "today"
			if len(args) > 0 {
				period = args[0]
			}

			var start, end time.Time
			now := time.Now()
			switch period {
			case "week":
				end = now
				start = now.AddDate(0, 0, -7)
			default: // today
				y, m, day := now.Date()
				start = time.Date(y, m, day, 0, 0, 0, 0, now.Location())
				end = start.Add(24 * time.Hour)
			}

			rows, err := d.QueryContext(cmd.Context(),
				`SELECT start_ts, end_ts, repo, focused_minutes, switches, tools, summary
				 FROM blocks WHERE start_ts >= ? AND start_ts < ? ORDER BY start_ts`,
				start.UTC(), end.UTC(),
			)
			if err != nil {
				return err
			}
			defer rows.Close()

			count := 0
			totalFocus := 0
			for rows.Next() {
				var startTS, endTS, repo, tools, summary string
				var focused, switches int
				rows.Scan(&startTS, &endTS, &repo, &focused, &switches, &tools, &summary)
				fmt.Printf("── %s → %s  repo:%-20s focus:%dm switches:%d\n",
					startTS, endTS, repo, focused, switches)
				totalFocus += focused
				count++
			}
			if count == 0 {
				fmt.Printf("no blocks for %s\n", period)
			} else {
				fmt.Printf("\ntotal: %d blocks, %d focused minutes\n", count, totalFocus)
			}
			return nil
		},
	}
}
