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

// onCmd represents the on command
var onCmd = &cobra.Command{
	Use:   "on [vm_name...]",
	Short: "Turn on the instances",
	Long: `Turn on the instances

Example:
  gcectl on [vm_name...]`,
	Args: cobra.MinimumNArgs(1),
	Run:  onRun,
}

func onRun(cmd *cobra.Command, args []string) {
	console := presenter.NewConsolePresenter()
	vmNames := args
	infraLog.DefaultLogger.Debugf("Turning on the instances %s", strings.Join(vmNames, ", "))

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
	startVMUseCase := usecase.NewStartVMUseCase(vmRepo, infraLog.DefaultLogger)

	// Turn on the instances
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = console.ExecuteWithProgress(
		ctx,
		fmt.Sprintf("Starting VMs %s", strings.Join(vmNames, ", ")),
		func(ctx context.Context) error {
			return startVMUseCase.Execute(ctx, vms)
		},
	)

	if err != nil {
		console.Error(fmt.Sprintf("Failed to turn on the instances: %v\n", err))
		os.Exit(1)
	}

	console.Success(fmt.Sprintf("Turned on the instances: %v\n", strings.Join(vmNames, ", ")))
}

func init() {
	rootCmd.AddCommand(onCmd)
}
