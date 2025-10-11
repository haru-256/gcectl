/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

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

// onCmd represents the on command
var onCmd = &cobra.Command{
	Use:   "on <vm_name>",
	Short: "Turn on the instance",
	Long: `Turn on the instance

Example:
  gcectl on <vm_name>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		console := presenter.NewConsolePresenter()
		vmName := args[0]
		infraLog.DefaultLogger.Debugf("Turning on the instance %s", vmName)
		if vmName == "" {
			console.Error("VM name is required")
			os.Exit(1)
		}

		// parse config
		cnf, err := config.ParseConfig(CnfPath)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to parse config: %v\n", err))
			os.Exit(1)
		}
		infraLog.DefaultLogger.Debug(fmt.Sprintf("Config: %+v", cnf))

		// filter VM by name
		vm := cnf.GetVMByName(vmName)
		if vm == nil {
			console.Error(fmt.Sprintf("VM %s not found", vmName))
			os.Exit(1)
		}

		// 依存性の注入
		vmRepo := gcp.NewVMRepository(CnfPath, infraLog.DefaultLogger)
		// Set progress callback to display dots during operation
		vmRepo.SetProgressCallback(console.Progress)
		startVMUseCase := usecase.NewStartVMUseCase(vmRepo)

		// Turn on the instance
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		console.ProgressStart(fmt.Sprintf("Starting VM %s", vmName))
		if err = startVMUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name); err != nil {
			console.ProgressDone()
			console.Error(fmt.Sprintf("Failed to turn on the instance: %v\n", err))
			os.Exit(1)
		}
		console.ProgressDone()
		console.Success(fmt.Sprintf("Turned on the instance: %v\n", vmName))
	},
}

func init() {
	rootCmd.AddCommand(onCmd)
}
