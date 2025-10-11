package set

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/haru-256/gcectl/internal/infrastructure/config"
	"github.com/haru-256/gcectl/internal/infrastructure/gcp"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
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
		console := presenter.NewConsolePresenter()
		vmName := args[0]
		machineType := args[1]
		if machineType == "" || vmName == "" {
			console.Error("machine-type and vm_name are required")
			os.Exit(1)
		}

		cnfPath, err := cmd.Flags().GetString("config")
		if err != nil {
			console.Error("config is required")
			os.Exit(1)
		}

		// parse config
		cnf, err := config.ParseConfig(cnfPath)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to parse config: %v\n", err))
			os.Exit(1)
		}
		infraLog.DefaultLogger.Debugf(fmt.Sprintf("Config: %+v", cnf))

		// filter VM by name
		vm := cnf.GetVMByName(vmName)
		if vm == nil {
			console.Error(fmt.Sprintf("VM %s not found", vmName))
			os.Exit(1)
		}

		// 依存性の注入
		vmRepo := gcp.NewVMRepository(cnfPath, infraLog.DefaultLogger)
		// Set progress callback to display dots during operation
		vmRepo.SetProgressCallback(console.Progress)
		updateMachineTypeUseCase := usecase.NewUpdateMachineTypeUseCase(vmRepo)

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		console.ProgressStart(fmt.Sprintf("Updating machine type for VM %s", vmName))
		if err = updateMachineTypeUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, machineType); err != nil {
			console.ProgressDone()
			console.Error(fmt.Sprintf("Failed to set machine-type: %v\n", err))
			os.Exit(1)
		}
		console.ProgressDone()
		console.Success(fmt.Sprintf("Set machine-type to %v\n", machineType))
	},
}

func init() {
	SetCmd.AddCommand(machineTypeCmd)
}
