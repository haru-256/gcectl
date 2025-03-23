package set

import (
	"os"

	"github.com/haru-256/gcectl/pkg/log"
	"github.com/haru-256/gcectl/pkg/utils"
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
		log.Logger.Debug("run root command")
		if err := cmd.Help(); err != nil {
			utils.ErrorReport("Failed to run help command")
			os.Exit(1)
		}
	},
}
