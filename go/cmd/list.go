package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/haru-256/gcectl/internal/infrastructure/gcp"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
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
		// 依存性の注入
		vmRepo := gcp.NewVMRepository(CnfPath, infraLog.DefaultLogger)
		console := presenter.NewConsolePresenter()
		listVMsUC := usecase.NewListVMsUseCase(vmRepo)

		// List VMs
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		items, err := listVMsUC.Execute(ctx)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to list VMs: %v\n", err))
			os.Exit(1)
		}

		infraLog.DefaultLogger.Debugf("Found %d VMs", len(items))

		// Convert usecase items to presenter items
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

		// Render VM list
		console.RenderVMList(presenterItems)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
