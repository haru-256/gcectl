package set

import (
	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/spf13/cobra"
)

var MachineTypeCmd = &cobra.Command{
	Use:   "machine-type",
	Short: "Set machine-type",
	Long:  `Set machine-type for the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Implement your logic here
		log.Logger.Debug("schedule called")
	},
}

func init() {
	SetCmd.AddCommand(MachineTypeCmd)
}
