package set

import (
	"os"

	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/spf13/cobra"
)

var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the machine type or schedule policy",
	Long:  `Set the machine type or schedule policy for the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger.Debug("run root command")
		if err := cmd.Help(); err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
	},
}
