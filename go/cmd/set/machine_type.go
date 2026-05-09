package set

import (
	"context"
	"fmt"
	"os"

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

		session, ctx, err := cli.NewSession(cmd, cnfPath)
		if err != nil {
			console.Error(err.Error())
			os.Exit(1)
		}
		defer session.Close()

		vm, err := session.Config.ResolveVM(vmName)
		if err != nil {
			session.Console.Error(err.Error())
			session.Close()
			os.Exit(1)
		}

		err = session.OpenVMRepository(ctx)
		if err != nil {
			session.Console.Error(err.Error())
			session.Close()
			os.Exit(1)
		}

		updateMachineTypeUseCase := usecase.NewUpdateMachineTypeUseCase(session.VMRepository, infraLog.DefaultLogger)

		message := fmt.Sprintf("Updating machine type for VM %s", vmName)
		err = session.Console.ExecuteWithProgress(ctx, message, func(ctx context.Context) error {
			return updateMachineTypeUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, machineType)
		})
		if err != nil {
			session.Console.Error(fmt.Sprintf("Failed to set machine-type: %v", err))
			session.Close()
			os.Exit(1)
		}
		session.Console.Success(fmt.Sprintf("Set machine-type to %v", machineType))
	},
}

func init() {
	SetCmd.AddCommand(machineTypeCmd)
}
