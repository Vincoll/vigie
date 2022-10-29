package vigiemain

import (
	"os"

	"github.com/vincoll/vigie/cmd/vigie/api"
	"github.com/vincoll/vigie/cmd/vigie/worker"

	"github.com/spf13/cobra"

	"github.com/vincoll/vigie/cmd/vigie/run"
	"github.com/vincoll/vigie/cmd/vigie/version"
)

var rootCmd = &cobra.Command{
	Use:   "webapi",
	Short: "webapi",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	}}

// Main entry
func Main() {

	addCommands()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	// Stay Alive for continuous monitoring
	stayAliveForEver()
}

// AddCommands adds child commands to the root command rootCmd.
func addCommands() {
	rootCmd.AddCommand(run.Cmd)
	rootCmd.AddCommand(vigieapi.Cmd)
	//	rootCmd.AddCommand(check.Cmd)
	rootCmd.AddCommand(worker.Cmd)
	//	rootCmd.AddCommand(check.Cmd)
	rootCmd.AddCommand(version.Cmd)

}

func stayAliveForEver() {
	select {}
}
