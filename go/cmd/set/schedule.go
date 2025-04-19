package set

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/haru-256/gcectl/pkg/config"
	"github.com/haru-256/gcectl/pkg/gce"
	"github.com/haru-256/gcectl/pkg/log"
	"github.com/haru-256/gcectl/pkg/utils"
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
		vmName := args[0]
		policyName := args[1]
		if policyName == "" || vmName == "" {
			utils.ErrorReport("schedule-policy and vm_name are required")
			os.Exit(1)
		}
		cnfPath, err := cmd.Flags().GetString("config")
		if err != nil {
			utils.ErrorReport("config is required")
			os.Exit(1)
		}
		// parse config
		cnf, err := config.ParseConfig(cnfPath)
		if err != nil {
			utils.ErrorReport(fmt.Sprintf("Failed to parse config: %v\n", err))
			os.Exit(1)
		}
		log.Logger.Debug(fmt.Sprintf("Config: %+v", cnf))
		// filter VM by name
		vm := cnf.GetVMByName(vmName)
		if vm == nil {
			utils.ErrorReport(fmt.Sprintf("VM %s not found", vmName))
			os.Exit(1)
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if unset {
			log.Logger.Debug("Unset schedule-policy")
			if err = gce.UnsetSchedulePolicy(ctx, vm, policyName); err != nil {
				fmt.Printf("Failed to unset schedule-policy: %v\n", err)
				utils.ErrorReport(fmt.Sprintf("Failed to unset schedule-policy: %v\n", err))
				os.Exit(1)
			} else {
				utils.SuccessReport(fmt.Sprintf("Unset schedule-policy: %v\n", policyName))
			}
		} else {
			log.Logger.Debug("Set schedule-policy")
			if err = gce.SetSchedulePolicy(ctx, vm, policyName); err != nil {
				utils.ErrorReport(fmt.Sprintf("Failed to set schedule-policy: %v\n", err))
				os.Exit(1)
			} else {
				utils.SuccessReport(fmt.Sprintf("Set schedule-policy: %v\n", policyName))
			}
		}
	},
}

var unset bool

func init() {
	SetCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().BoolVarP(&unset, "un", "u", false, "Unset schedule-policy")
}
