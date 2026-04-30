package fw

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgPath string
	debug   bool
)

func Run() {
	root := &cobra.Command{
		Use:   "fw",
		Short: "flowd — local coding activity daemon",
		PersistentPreRun: func(*cobra.Command, []string) {
			InitLogger(debug)
		},
	}
	root.PersistentFlags().StringVar(&cfgPath, "config", DefaultConfigPath(), "config file path")
	root.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	root.AddCommand(cmdInit(), cmdStart(), cmdStatus(), cmdSummary(), cmdReport(), cmdDashboard(), cmdSetupTmux())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func cmdInit() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup",
		RunE: func(*cobra.Command, []string) error {
			if _, err := os.Stat(cfgPath); err == nil && !force {
				fmt.Printf("config already exists at %s (use --force to overwrite)\n", cfgPath)
				return nil
			}
			cfg, err := RunInitWizard()
			if err != nil {
				return err
			}
			if err := WriteConfig(cfgPath, cfg); err != nil {
				return err
			}
			fmt.Printf("\n  config → %s\n", cfgPath)

			d, err := OpenDB(cfg.DBPath)
			if err != nil {
				return err
			}
			d.Close()
			fmt.Printf("  db ready → %s\n", cfg.DBPath)

			if err := SetupRepo(cfg.RepoPath, cfg.GitRemote, cfg.Branch); err != nil {
				fmt.Printf("  warn: repo setup: %v\n", err)
			}
			if AskTmuxAutostart() {
				if err := SetupTmuxAutostart(); err != nil {
					fmt.Printf("  warn: tmux autostart: %v\n", err)
				}
			}
			fmt.Println("\n  run `fw start` to begin tracking")
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite existing config")
	return c
}

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

			fmt.Println("flowd started (ctrl+c to stop)")
			slog.Info("daemon up", "db", cfg.DBPath, "poll_sec", cfg.PollIntervalSec)

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
		if cfg.GitRemote != "" {
			// flush WAL so the on-disk .db is current before commit
			if _, err := d.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
				slog.Warn("wal checkpoint", "err", err)
			}
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

func cmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show DB status",
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
			var ev, bl int
			d.QueryRow("SELECT COUNT(*) FROM events").Scan(&ev)
			d.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&bl)
			fmt.Printf("db:     %s\nevents: %d\nblocks: %d\n", cfg.DBPath, ev, bl)
			return nil
		},
	}
}

func cmdSummary() *cobra.Command {
	var save bool
	c := &cobra.Command{
		Use:   "summary",
		Short: "Build and print the current block (use --save to persist it)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			d, err := OpenDB(cfg.DBPath)
			if err != nil {
				return err
			}
			defer d.Close()
			now := time.Now().UTC()
			iv := time.Duration(cfg.SummaryIntervalMin) * time.Minute
			u := now.UnixNano()
			// align end to last boundary; use now as end if --save so we
			// capture all events up to this moment
			var start, end time.Time
			if save {
				end = now
				start = time.Unix(0, u-(u%iv.Nanoseconds())).UTC()
				if start.Equal(end) || start.After(end) {
					start = end.Add(-iv)
				}
			} else {
				end = time.Unix(0, u-(u%iv.Nanoseconds())).UTC()
				start = end.Add(-iv)
			}
			b, err := BuildBlock(cmd.Context(), d, start, end, cfg.PollIntervalSec, save)
			if err != nil {
				return err
			}
			if save {
				fmt.Println("block saved.")
			}
			fmt.Println(b.Summary)
			return nil
		},
	}
	c.Flags().BoolVar(&save, "save", false, "persist the block to the DB (shows in dashboard)")
	return c
}

func cmdReport() *cobra.Command {
	return &cobra.Command{
		Use:   "report [today|week]",
		Short: "Text activity report",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			d, err := OpenDB(cfg.DBPath)
			if err != nil {
				return err
			}
			defer d.Close()
			period := "today"
			if len(args) > 0 {
				period = args[0]
			}
			start, end := PeriodRange(period, time.Now())
			blocks, err := LoadBlocks(cmd.Context(), d, start, end)
			if err != nil {
				return err
			}
			fmt.Print(TextReport(blocks, period))
			return nil
		},
	}
}

func cmdDashboard() *cobra.Command {
	var noOpen, aiRecap bool
	var outFlag string
	c := &cobra.Command{
		Use:   "dashboard [today|week]",
		Short: "Render an HTML dashboard and open it in the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			d, err := OpenDB(cfg.DBPath)
			if err != nil {
				return err
			}
			defer d.Close()
			period := "today"
			if len(args) > 0 {
				period = args[0]
			}
			start, end := PeriodRange(period, time.Now())
			blocks, err := LoadBlocks(cmd.Context(), d, start, end)
			if err != nil {
				return err
			}

			recap := ""
			if aiRecap {
				if !cfg.AIEnabled || cfg.AICommand == "" {
					fmt.Println("note: --ai-recap requires ai_enabled + ai_command in config; skipping")
				} else {
					recap, err = runAIRecap(cmd.Context(), cfg, blocks, period)
					if err != nil {
						fmt.Printf("note: ai recap failed: %v\n", err)
					}
				}
			}

			out := outFlag
			if out == "" {
				out = filepath.Join(os.TempDir(), fmt.Sprintf("flowd-%s.html", period))
			}
			if err := RenderDashboard(blocks, period, recap, out); err != nil {
				return err
			}
			fmt.Printf("dashboard → %s\n", out)
			if !noOpen {
				if err := OpenInBrowser(out); err != nil {
					fmt.Printf("(could not open browser: %v — open the file manually)\n", err)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&noOpen, "no-open", false, "do not open the browser")
	c.Flags().BoolVar(&aiRecap, "ai-recap", false, "run AI on all blocks for an aggregate recap (slow)")
	c.Flags().StringVar(&outFlag, "out", "", "output file path (default: temp dir)")
	return c
}

func runAIRecap(ctx context.Context, cfg *Config, blocks []Block, period string) (string, error) {
	if len(blocks) == 0 {
		return "", nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Period: %s\n\n", period)
	for _, bl := range blocks {
		b.WriteString(bl.Summary)
		b.WriteString("\n")
	}
	prompt := cfg.AIPrompt
	if prompt == "" {
		prompt = "Give me a 3-bullet recap of this coding period: what was worked on, patterns noticed, and one suggestion. Be concise."
	} else {
		prompt = "Aggregate recap of multiple coding blocks. " + prompt
	}
	return RunAI(ctx, cfg.AICommand, prompt, b.String())
}

func cmdSetupTmux() *cobra.Command {
	return &cobra.Command{
		Use:   "setup-tmux",
		Short: "Add `fw start` to ~/.tmux.conf",
		RunE: func(*cobra.Command, []string) error {
			return SetupTmuxAutostart()
		},
	}
}
