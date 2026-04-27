package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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
		cmdSetupTmux(),
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

			// offer tmux autostart
			if initwizard.AskTmuxAutostart() {
				if err := setupTmuxAutostart(); err != nil {
					fmt.Printf("  warning: tmux autostart setup failed: %v\n", err)
					fmt.Println("  add manually to ~/.tmux.conf:")
					fmt.Println(`    run-shell "fw start &> /tmp/flowd.log &"`)
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
				tracker := session.NewTracker(d, cfg.PollIntervalSec, cfg.WatchDirs)
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

func cmdSetupTmux() *cobra.Command {
	return &cobra.Command{
		Use:   "setup-tmux",
		Short: "Add fw start to ~/.tmux.conf so it runs when tmux starts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setupTmuxAutostart()
		},
	}
}

// setupTmuxAutostart appends a run-shell line to ~/.tmux.conf.
// Idempotent — skips if the line already exists.
func setupTmuxAutostart() error {
	home, _ := os.UserHomeDir()
	confPath := home + "/.tmux.conf"

	const marker = "fw start"

	// check if already present
	existing, err := os.ReadFile(confPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read tmux.conf: %w", err)
	}
	if strings.Contains(string(existing), marker) {
		fmt.Println("  tmux autostart already configured in ~/.tmux.conf")
		return nil
	}

	f, err := os.OpenFile(confPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open tmux.conf: %w", err)
	}
	defer f.Close()

	line := "\n# flowd — start activity tracker with tmux\nrun-shell \"fw start &> /tmp/flowd.log &\"\n"
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write tmux.conf: %w", err)
	}

	fmt.Println("  added to ~/.tmux.conf:")
	fmt.Println(`    run-shell "fw start &> /tmp/flowd.log &"`)
	fmt.Println("  flowd will start automatically when tmux starts")

	// reload tmux config if a server is running
	if err := exec.Command("tmux", "source-file", confPath).Run(); err == nil {
		fmt.Println("  tmux config reloaded")
	}
	return nil
}

// confirmPrompt is used by init to ask about tmux autostart inline.
func confirmPrompt(prompt string) bool {
	fmt.Printf("  %s (y/n) [n]: ", prompt)
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(line)) == "y"
}
