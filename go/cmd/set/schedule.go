package set

import (
	"context"
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

var scheduleCmd = &cobra.Command{
	Use:   "schedule-policy <vm_name> <policy_name>",
	Short: "Set schedule-policy",
	Long: `Set schedule-policy for the application.

Example:
  gcectl set schedule-policy sandbox stop`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		console := presenter.NewConsolePresenter()
		vmName := args[0]
		policyName := args[1]
		if policyName == "" || vmName == "" {
			console.Error("schedule-policy and vm_name are required")
			os.Exit(1)
		}

		cnfPath, err := cmd.Flags().GetString("config")
		if err != nil {
			console.Error("config is required")
			os.Exit(1)
		}

		// parse config
		cnf, err := config.ParseConfig(cnfPath)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to parse config: %v\n", err))
			os.Exit(1)
		}
		infraLog.DefaultLogger.Debugf(fmt.Sprintf("Config: %+v", cnf))

		// filter VM by name
		vm := cnf.GetVMByName(vmName)
		if vm == nil {
			console.Error(fmt.Sprintf("VM %s not found", vmName))
			os.Exit(1)
		}

		// 依存性の注入
		vmRepo := gcp.NewVMRepository(cnfPath, infraLog.DefaultLogger)

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		if unset {
			infraLog.DefaultLogger.Debugf("Unset schedule-policy")
			unsetSchedulePolicyUseCase := usecase.NewUnsetSchedulePolicyUseCase(vmRepo, infraLog.DefaultLogger)

			var message string
			if vm.SchedulePolicy != "" {
				message = fmt.Sprintf("Unsetting schedule policy %s for VM %s", vm.SchedulePolicy, vmName)
			} else {
				message = fmt.Sprintf("Unsetting schedule policy for VM %s", vmName)
			}

			err = console.ExecuteWithProgress(ctx, message, func(ctx context.Context) error {
				return unsetSchedulePolicyUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, policyName)
			})

			if err != nil {
				console.Error(fmt.Sprintf("Failed to unset schedule-policy: %v\n", err))
				os.Exit(1)
			}
			console.Success(fmt.Sprintf("Unset schedule-policy: %v\n", policyName))
		} else {
			infraLog.DefaultLogger.Debugf("Set schedule-policy")
			setSchedulePolicyUseCase := usecase.NewSetSchedulePolicyUseCase(vmRepo, infraLog.DefaultLogger)

			message := fmt.Sprintf("Setting schedule policy %s for VM %s", policyName, vmName)

			err = console.ExecuteWithProgress(ctx, message, func(ctx context.Context) error {
				return setSchedulePolicyUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, policyName)
			})

			if err != nil {
				console.Error(fmt.Sprintf("Failed to set schedule-policy: %v\n", err))
				os.Exit(1)
			}
			console.Success(fmt.Sprintf("Set schedule-policy: %v\n", policyName))
		}
	},
}

var unset bool

func init() {
	SetCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().BoolVarP(&unset, "un", "u", false, "Unset schedule-policy")
}
