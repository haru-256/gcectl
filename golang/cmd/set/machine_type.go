package set

import (
	"fmt"
	"os"

	"github.com/haru-256/gce-commands/pkg/config"
	"github.com/haru-256/gce-commands/pkg/gce"
	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/spf13/cobra"
)

var machineTypeCmd = &cobra.Command{
	Use:   "machine-type <vm_name> <machine-type>",
	Short: "Set machine-type",
	Long: `Set machine-type for the application.

Example:
  gce-commands set machine-type sandbox n1-standard-1`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		machineType := args[1]
		if machineType == "" || vmName == "" {
			log.Logger.Error("machine-type and vm_name are required")
			os.Exit(1)
		}
		cnfPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
		// parse config
		cnf, err := config.ParseConfig(cnfPath)
		if err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
		log.Logger.Debug(fmt.Sprintf("Config: %+v", cnf))
		// filter VM by name
		vm := cnf.GetVMByName(vmName)
		// Implement your logic here
		if err = gce.SetMachineType(vm, machineType); err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	SetCmd.AddCommand(machineTypeCmd)
}
