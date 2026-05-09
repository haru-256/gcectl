/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
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
		vmName := args[0]
		infraLog.DefaultLogger.Debugf("Describe instance %s", vmName)
		if vmName == "" {
			presenter.NewConsolePresenter().Error("VM name is required")
			os.Exit(1)
		}

		session, ctx, err := cli.NewSession(cmd, CnfPath)
		if err != nil {
			presenter.NewConsolePresenter().Error(err.Error())
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

		describeVMUseCase := usecase.NewDescribeVMUseCase(session.VMRepository)

		vmDetail, uptimeStr, err := describeVMUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name)
		if err != nil {
			session.Console.Error(fmt.Sprintf("Failed to get VM info: %v", err))
			session.Close()
			os.Exit(1)
		}

		session.Console.RenderVMDetail(presenter.VMDetail{
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
