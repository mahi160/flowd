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
	"github.com/mahi/flowd/internal/initwizard"
	"github.com/mahi/flowd/internal/logger"
	"github.com/mahi/flowd/internal/report"
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
		cmdInit(),
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

func cmdInit() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup — create config and initialize DB",
		RunE: func(cmd *cobra.Command, args []string) error {
			dest := cfgPath
			if _, err := os.Stat(dest); err == nil && !force {
				fmt.Printf("config already exists at %s\n", dest)
				fmt.Println("run with --force to overwrite")
				return nil
			}

			cfg, err := initwizard.Run()
			if err != nil {
				return err
			}

			if err := config.Write(dest, cfg); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
			fmt.Printf("\n  config written → %s\n", dest)

			// ensure DB is created and migrated
			d, err := db.Open(cfg.DBPath)
			if err != nil {
				return fmt.Errorf("init db: %w", err)
			}
			d.Close()
			fmt.Printf("  database ready → %s\n", cfg.DBPath)

			// set up private git repo if sync enabled
			if cfg.PushDB {
				if err := initwizard.SetupRepo(cfg.RepoPath, cfg.GitRemote, cfg.Branch); err != nil {
					fmt.Printf("  warning: repo setup failed: %v\n", err)
					fmt.Println("  you can set it up manually — see README for instructions")
				}
			}

			fmt.Println()
			fmt.Println("  run `fw start` to begin tracking")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing config")
	return cmd
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
	var htmlOut bool
	cmd := &cobra.Command{
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
			default:
				y, m, day := now.Date()
				start = time.Date(y, m, day, 0, 0, 0, 0, now.Location())
				end = start.Add(24 * time.Hour)
			}

			blocks, err := report.QueryBlocks(cmd.Context(), d, start, end)
			if err != nil {
				return err
			}

			if htmlOut {
				fmt.Print(report.HTMLReport(blocks, period))
			} else {
				fmt.Print(report.TextReport(blocks, period))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&htmlOut, "html", false, "output HTML dashboard")
	return cmd
}
