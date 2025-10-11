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

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe <vm_name>",
	Short: "Describe the instance",
	Long: `Describe the instance.

Example:
  gcectl describe <vm_name>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		console := presenter.NewConsolePresenter()
		vmName := args[0]
		infraLog.DefaultLogger.Debugf("Describe instance %s", vmName)
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

		// Describe the instance
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		vmDetail, uptimeStr, err := usecase.DescribeVM(ctx, vmRepo, vm.Project, vm.Zone, vm.Name)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to get VM info: %v\n", err))
			os.Exit(1)
		}

		// Render VM detail
		console.RenderVMDetail(presenter.VMDetail{
			Name:           vmDetail.Name,
			Project:        vmDetail.Project,
			Zone:           vmDetail.Zone,
			MachineType:    vmDetail.MachineType,
			Status:         vmDetail.Status,
			SchedulePolicy: vmDetail.SchedulePolicy,
			Uptime:         uptimeStr,
		})
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
