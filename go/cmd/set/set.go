package set

import (
	"os"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/spf13/cobra"
)

var SetCmd = &cobra.Command{
	Use:   "set <command>",
	Short: "Set the machine type or schedule policy",
	Long: `Set the machine type or schedule policy for the application.

Example:
  gcectl set machine-type sandbox n1-standard-1
  gcectl set schedule-policy sandbox stop`,
	Run: func(cmd *cobra.Command, args []string) {
		console := presenter.NewConsolePresenter()
		infraLog.DefaultLogger.Debugf("run root command")
		if err := cmd.Help(); err != nil {
			console.Error("Failed to run help command")
			os.Exit(1)
		}
	},
}
