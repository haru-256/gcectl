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

		if unset {
			infraLog.DefaultLogger.Debugf("Unset schedule-policy")
			unsetSchedulePolicyUseCase := usecase.NewUnsetSchedulePolicyUseCase(session.VMRepository, infraLog.DefaultLogger)

			var message string
			if vm.SchedulePolicy != "" {
				message = fmt.Sprintf("Unsetting schedule policy %s for VM %s", vm.SchedulePolicy, vmName)
			} else {
				message = fmt.Sprintf("Unsetting schedule policy for VM %s", vmName)
			}

			err = session.Console.ExecuteWithProgress(ctx, message, func(ctx context.Context) error {
				return unsetSchedulePolicyUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, policyName)
			})

			if err != nil {
				session.Console.Error(fmt.Sprintf("Failed to unset schedule-policy: %v", err))
				session.Close()
				os.Exit(1)
			}
			session.Console.Success(fmt.Sprintf("Unset schedule-policy: %v", policyName))
		} else {
			infraLog.DefaultLogger.Debugf("Set schedule-policy")
			setSchedulePolicyUseCase := usecase.NewSetSchedulePolicyUseCase(session.VMRepository, infraLog.DefaultLogger)

			message := fmt.Sprintf("Setting schedule policy %s for VM %s", policyName, vmName)

			err = session.Console.ExecuteWithProgress(ctx, message, func(ctx context.Context) error {
				return setSchedulePolicyUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, policyName)
			})

			if err != nil {
				session.Console.Error(fmt.Sprintf("Failed to set schedule-policy: %v", err))
				session.Close()
				os.Exit(1)
			}
			session.Console.Success(fmt.Sprintf("Set schedule-policy: %v", policyName))
		}
	},
}

var unset bool

func init() {
	SetCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().BoolVarP(&unset, "un", "u", false, "Unset schedule-policy")
}
