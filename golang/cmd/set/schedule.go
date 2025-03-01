package set

import (
	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Set schedule policy",
	Long:  `Set schedule policy for the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Implement your logic here
		log.Logger.Debug("schedule called")
	},
}

func init() {
	SetCmd.AddCommand(scheduleCmd)
}
