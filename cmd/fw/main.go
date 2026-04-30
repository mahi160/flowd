package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mahi/flowd/internal/config"
	"github.com/mahi/flowd/internal/db"
	"github.com/mahi/flowd/internal/logger"
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

			fmt.Println("flowd daemon starting (foreground mode)")
			logger.L.Info("daemon started", "db", cfg.DBPath, "poll_sec", cfg.PollIntervalSec)

			// Phase 2 will wire collectors here
			select {}
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
	cmd := &cobra.Command{
		Use:   "summary [now]",
		Short: "Generate a summary for the current or last 30-min block",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("summary not yet implemented (Phase 3)")
			return nil
		},
	}
	return cmd
}

func cmdReport() *cobra.Command {
	return &cobra.Command{
		Use:   "report [today|week]",
		Short: "Show activity report",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("report not yet implemented (Phase 5)")
			return nil
		},
	}
}
