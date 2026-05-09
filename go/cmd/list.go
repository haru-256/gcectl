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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VM in settings",
	Long: `List all VM in settings.

Example:
  gcectl list`,
	Run: func(cmd *cobra.Command, args []string) {
		console := presenter.NewConsolePresenter()
		session, ctx, err := cli.NewSession(cmd, CnfPath)
		if err != nil {
			console.Error(err.Error())
			os.Exit(1)
		}
		defer session.Close()

		err = session.OpenVMRepository(ctx)
		if err != nil {
			console.Error(err.Error())
			session.Close()
			os.Exit(1)
		}

		listVMsUC := usecase.NewListVMsUseCase(session.VMRepository)

		items, err := listVMsUC.Execute(ctx, session.Config.VMs)
		infraLog.DefaultLogger.Debugf("Found %d VMs", len(items))

		presenterItems := make([]presenter.VMListItem, len(items))
		for i, item := range items {
			presenterItems[i] = presenter.VMListItem{
				Name:           item.VM.Name,
				Project:        item.VM.Project,
				Zone:           item.VM.Zone,
				MachineType:    item.VM.MachineType,
				Status:         item.VM.Status,
				SchedulePolicy: item.VM.SchedulePolicy,
				Uptime:         item.Uptime,
			}
		}

		if len(presenterItems) > 0 {
			console.RenderVMList(presenterItems)
		}
		if err != nil {
			console.Error(fmt.Sprintf("Failed to list some VMs: %v", err))
			session.Close()
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
