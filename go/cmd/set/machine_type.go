package set

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/haru-256/gcectl/pkg/config"
	"github.com/haru-256/gcectl/pkg/gce"
	"github.com/haru-256/gcectl/pkg/log"
	"github.com/haru-256/gcectl/pkg/utils"
	"github.com/spf13/cobra"
)

var machineTypeCmd = &cobra.Command{
	Use:   "machine-type <vm_name> <machine-type>",
	Short: "Set machine-type",
	Long: `Set machine-type for the application.

Example:
  gcectl set machine-type sandbox n1-standard-1`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		machineType := args[1]
		if machineType == "" || vmName == "" {
			utils.ErrorReport("machine-type and vm_name are required")
			os.Exit(1)
		}
		cnfPath, err := cmd.Flags().GetString("config")
		if err != nil {
			utils.ErrorReport("config is required")
			os.Exit(1)
		}
		// parse config
		cnf, err := config.ParseConfig(cnfPath)
		if err != nil {
			utils.ErrorReport(fmt.Sprintf("Failed to parse config: %v\n", err))
			os.Exit(1)
		}
		log.Logger.Debug(fmt.Sprintf("Config: %+v", cnf))
		// filter VM by name
		vm := cnf.GetVMByName(vmName)
		if vm == nil {
			utils.ErrorReport(fmt.Sprintf("VM %s not found", vmName))
			os.Exit(1)
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err = gce.SetMachineType(ctx, vm, machineType); err != nil {
			utils.ErrorReport(fmt.Sprintf("Failed to set machine-type: %v\n", err))
			os.Exit(1)
		}
		utils.SuccessReport(fmt.Sprintf("Set machine-type: %v\n", machineType))
	},
}

func init() {
	SetCmd.AddCommand(machineTypeCmd)
}
