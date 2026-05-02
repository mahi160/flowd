package fw

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via:
//
//	go build -ldflags "-X github.com/mahi160/flowd/internal/fw.Version=1.2.3" ./cmd/fw
//
// Falls back to "dev" when building without the flag.
var Version = "dev"

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
	root.Version = Version
	root.SetVersionTemplate(fmt.Sprintf("fw %s\n", Version))
	root.AddCommand(
		cmdInit(), cmdStart(), cmdStop(), cmdStatus(),
		cmdSummary(), cmdReport(), cmdDashboard(), cmdSetupTmux(),
		cmdUpdate(), cmdScanAI(),
	)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
