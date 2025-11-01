package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/haru-256/gcectl/internal/domain/model"
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

	// parse config
	cnf, err := config.ParseConfig(CnfPath)
	if err != nil {
		console.Error(fmt.Sprintf("Failed to parse config: %v\n", err))
		os.Exit(1)
	}
	infraLog.DefaultLogger.Debug(fmt.Sprintf("Config: %+v", cnf))

	// domain entity変換
	var vms []*model.VM
	for _, vmName := range vmNames {
		vm := cnf.GetVMByName(vmName)
		if vm == nil {
			console.Error(fmt.Sprintf("VM %s not found", vmName))
			os.Exit(1)
		}
		vms = append(vms, vm)
	}

	// 依存性の注入
	vmRepo := gcp.NewVMRepository(CnfPath, infraLog.DefaultLogger)
	stopVMUseCase := usecase.NewStopVMUseCase(vmRepo, infraLog.DefaultLogger)

	// Turn off the instances
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = console.ExecuteWithProgress(ctx,
		fmt.Sprintf("Starting VMs %s", strings.Join(vmNames, ", ")),
		func(ctx context.Context) error {
			return stopVMUseCase.Execute(ctx, vms)
		})
	if err != nil {
		console.Error(fmt.Sprintf("Failed to turn off the instance(s): %v\n", err))
		os.Exit(1)
	}

	console.Success(fmt.Sprintf("Turned on the instances: %v\n", strings.Join(vmNames, ", ")))
}

func init() {
	rootCmd.AddCommand(offCmd)
}
