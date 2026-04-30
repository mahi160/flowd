package fw

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

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

			// Determine blockStart: resume from persisted state if available.
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
				// Force-close the block: persist the new blockStart.
				if err := d.SetState(stateBlockStart, end.UTC().Format(time.RFC3339)); err != nil {
					return fmt.Errorf("persist block start: %w", err)
				}
				if err := WriteJournal(cfg, b); err != nil {
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
