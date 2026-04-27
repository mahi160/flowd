package fw

import (
	"os"

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
	root.AddCommand(
		cmdInit(), cmdStart(), cmdStop(), cmdStatus(),
		cmdSummary(), cmdReport(), cmdDashboard(), cmdSetupTmux(),
	)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
