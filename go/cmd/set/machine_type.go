package set

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/haru-256/gcectl/internal/infrastructure/gcp"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
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

		vm, err := cli.ResolveVMByName(cnfPath, vmName)
		if err != nil {
			console.Error(fmt.Sprintf("%v\n", err))
			os.Exit(1)
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// 依存性の注入
		vmRepo, err := gcp.NewVMRepository(ctx, infraLog.DefaultLogger)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to create VM repository: %v\n", err))
			os.Exit(1)
		}
		defer func() {
			_ = vmRepo.Close()
		}()
		updateMachineTypeUseCase := usecase.NewUpdateMachineTypeUseCase(vmRepo, infraLog.DefaultLogger)

		message := fmt.Sprintf("Updating machine type for VM %s", vmName)
		err = console.ExecuteWithProgress(ctx, message, func(ctx context.Context) error {
			return updateMachineTypeUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, machineType)
		})

		if err != nil {
			console.Error(fmt.Sprintf("Failed to set machine-type: %v\n", err))
			os.Exit(1)
		}
		console.Success(fmt.Sprintf("Set machine-type to %v\n", machineType))
	},
}

func init() {
	SetCmd.AddCommand(machineTypeCmd)
}
