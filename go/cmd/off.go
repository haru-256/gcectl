package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/haru-256/gcectl/internal/infrastructure/config"
	"github.com/haru-256/gcectl/internal/infrastructure/gcp"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)

// offCmd represents the off command
var offCmd = &cobra.Command{
	Use:   "off <vm_name>...",
	Short: "Turn off one or more instances",
	Long: `Turn off one or more instances

Example:
  gcectl off <vm_name>
  gcectl off <vm_name1> <vm_name2> <vm_name3>`,
	Args: cobra.MinimumNArgs(1),
	Run:  offRun,
}

func offRun(cmd *cobra.Command, args []string) {
	console := presenter.NewConsolePresenter()
	vmNames := args
	infraLog.DefaultLogger.Debugf("Turning off the instances %s", strings.Join(vmNames, ", "))

	cfg, err := config.NewConfig(CnfPath)
	if err != nil {
		console.Error(err.Error())
		os.Exit(1)
	}

	vms, err := cfg.ResolveVMs(vmNames)
	if err != nil {
		console.Error(err.Error())
		os.Exit(1)
	}

	// Turn off the instances
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 依存性の注入
	vmRepo, err := gcp.NewVMRepository(ctx, infraLog.DefaultLogger)
	if err != nil {
		console.Error(fmt.Sprintf("Failed to create VM repository: %v", err))
		os.Exit(1)
	}
	defer func() {
		_ = vmRepo.Close()
	}()
	stopVMUseCase := usecase.NewStopVMUseCase(vmRepo, infraLog.DefaultLogger)

	err = console.ExecuteWithProgress(ctx,
		fmt.Sprintf("Stopping VMs %s", strings.Join(vmNames, ", ")),
		func(ctx context.Context) error {
			return stopVMUseCase.Execute(ctx, vms)
		})
	if err != nil {
		console.Error(fmt.Sprintf("Failed to turn off the instance(s): %v", err))
		os.Exit(1)
	}

	console.Success(fmt.Sprintf("Turned off the instances: %v", strings.Join(vmNames, ", ")))
}

func init() {
	rootCmd.AddCommand(offCmd)
}
