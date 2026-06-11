package fw

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/mahi160/flowd/internal/ai_sessions"
	"github.com/spf13/cobra"
)

func cmdScanAI() *cobra.Command {
	return &cobra.Command{
		Use:   "scan-ai",
		Short: "Manually trigger AI session scan",
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
			svc := ai_sessions.NewService(d.DB, cfg.AISessionPaths)
			return svc.RunSync()
		},
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

			d, err := OpenDB(cfg.DBPath())
			if err != nil {
				return err
			}
			d.Close()
			fmt.Printf("  db ready → %s\n", cfg.DBPath())

			if err := SetupRepo(cfg.RepoPath, cfg.GitRemote, cfg.Branch); err != nil {
				fmt.Printf("  warn: repo setup: %v\n", err)
			}
			tmuxAutostart := AskTmuxAutostart()
			if tmuxAutostart {
				if err := SetupTmuxAutostart(); err != nil {
					fmt.Printf("  warn: tmux autostart: %v\n", err)
				}
			}
			printReadyBanner(cfg.DBPath(), tmuxAutostart)
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite existing config")
	return c
}

func printReadyBanner(dbPath string, tmuxAutostart bool) {
	const (
		reset  = "\033[0m"
		bold   = "\033[1m"
		dim    = "\033[2m"
		green  = "\033[32m"
		cyan   = "\033[36m"
		indigo = "\033[94m"
	)

	startLine := "  " + bold + cyan + "fw start" + reset + dim + "            start the daemon" + reset
	if tmuxAutostart {
		startLine = "  " + dim + "fw start" + reset + dim + "            (runs automatically with tmux ✓)" + reset
	}

	fmt.Print(`
` + bold + indigo + `
  ██████╗  ██╗      ██████╗ ██╗    ██╗██████╗
  ██╔════╝ ██║     ██╔═══██╗██║    ██║██╔══██╗
  █████╗   ██║     ██║   ██║██║ █╗ ██║██║  ██║
  ██╔══╝   ██║     ██║   ██║██║███╗██║██║  ██║
  ██║      ███████╗╚██████╔╝╚███╔███╔╝██████╔╝
  ╚═╝      ╚══════╝ ╚═════╝  ╚══╝╚══╝ ╚═════╝` + reset + `
`)

	fmt.Println("  " + dim + "───────────────────────────────────────────────" + reset)
	fmt.Println(startLine)
	fmt.Println("  " + bold + cyan + "fw dashboard" + reset + dim + "        open the dashboard" + reset)
	fmt.Println("  " + bold + cyan + "fw report today" + reset + dim + "     today’s activity" + reset)
	fmt.Println("  " + bold + cyan + "fw status" + reset + dim + "           daemon status" + reset)
	fmt.Println("  " + dim + "───────────────────────────────────────────────" + reset)
	fmt.Println("  " + dim + "logs  ~/.local/share/flowd/flowd.log" + reset)
	fmt.Println("  " + dim + "db    " + dbPath + reset)
	fmt.Println()
}

func cmdSummary() *cobra.Command {
	var save bool
	c := &cobra.Command{
		Use:   "summary",
		Short: "Print the current in-progress block (since last completed block)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			d, err := OpenDB(cfg.DBPath())
			if err != nil {
				return err
			}
			defer d.Close()

			blockStart := time.Now().Add(-time.Duration(cfg.FocusBlockMin) * time.Minute)
			if v := d.GetState(stateBlockStart); v != "" {
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					blockStart = t
				}
			}
			end := time.Now().UTC()

			b, err := BuildBlock(cmd.Context(), d, blockStart, end, cfg.PollIntervalSec, save)
			if err != nil {
				return err
			}
			if save {
				if err := d.SetState(stateBlockStart, end.UTC().Format(time.RFC3339)); err != nil {
					return fmt.Errorf("persist block start: %w", err)
				}
				if err := WriteJournal(cmd.Context(), cfg, d, b); err != nil {
					return fmt.Errorf("journal write: %w", err)
				}
				fmt.Println("block saved.")
			}
			fmt.Println(b.Summary)
			return nil
		},
	}
	c.Flags().BoolVar(&save, "save", false, "force-close the current block and write it to the journal")
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
			d, err := OpenDB(cfg.DBPath())
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
	var noOpen bool
	var outFlag string
	c := &cobra.Command{
		Use:   "dashboard [today|week|month|year|all]",
		Short: "Build a full interactive dashboard (all period tabs) and open it",
		Long: `Loads data for all five periods (today, week, month, year, all) in one
pass so every tab in the dashboard works immediately. The optional period
argument sets which tab is shown first (default: today).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			d, err := OpenDB(cfg.DBPath())
			if err != nil {
				return err
			}
			defer d.Close()
			initPlatform(cfg.MachineName)

			// The optional arg sets the initial tab, not which data to load.
			initialPeriod := "today"
			if len(args) > 0 {
				valid := map[string]bool{"today": true, "week": true, "month": true, "year": true, "all": true}
				if !valid[args[0]] {
					return fmt.Errorf("invalid period %q (want today, week, month, year, or all)", args[0])
				}
				initialPeriod = args[0]
			}

			// Load blocks and AI rows for every period.
			allPeriods := []string{"today", "yesterday", "week", "month", "year", "all"}
			allBlocks := map[string][]Block{}
			allAIRows := map[string][]aiRawRow{}
			for _, p := range allPeriods {
				start, end := PeriodRange(p, time.Now())
				blocks, err := LoadBlocks(cmd.Context(), d, start, end)
				if err != nil {
					return fmt.Errorf("load blocks for %s: %w", p, err)
				}
				allBlocks[p] = blocks
				rows, err := loadAIRows(cmd.Context(), d, start, end)
				if err != nil {
					slog.Warn("load ai sessions", "period", p, "err", err)
				}
				allAIRows[p] = rows
			}

			// Streak uses all historical blocks.
			streakDays := d.QueryStreak(cmd.Context())

			payload := buildDashPayload(initialPeriod, allBlocks, allAIRows, streakDays)

			// Build today/yesterday standup (replaces the old --ai-recap flag).
			standup, err := BuildStandup(cmd.Context(), cfg,
				allBlocks["today"], allBlocks["yesterday"])
			if err != nil {
				slog.Warn("standup build", "err", err)
			} else if standup != nil {
				payload.Standup = standup.Text
				payload.StandupRaw = standup.Raw
			}

			out := outFlag
			if out == "" {
				out = filepath.Join(os.TempDir(), "flowd.html")
			}
			if err := RenderDashboard(payload, out); err != nil {
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
	c.Flags().StringVar(&outFlag, "out", "", "output file path (default: $TMPDIR/flowd.html)")
	return c
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

func cmdSetupNvim() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "setup-nvim",
		Short: "Install the flowd.lua neovim plugin",
		Long: `Writes the bundled flowd.lua to ~/.config/nvim/plugin/flowd.lua.

The plugin reports your active filetype to flowd on every buffer switch,
giving accurate language attribution before a git commit lands.
It works with or without a plugin manager and is fully optional.`,
		RunE: func(*cobra.Command, []string) error {
			if NvimPluginInstalled() && !force {
				fmt.Printf("already installed: %s/plugin/flowd.lua\n", nvimConfigDir())
				fmt.Println("run with --force to overwrite")
				return nil
			}
			dest, err := InstallNvimPlugin()
			if err != nil {
				return err
			}
			fmt.Printf("installed → %s\n", dest)
			fmt.Println("restart nvim (or :source the file) to activate.")
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite existing plugin file")
	return c
}
