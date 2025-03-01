package set

import (
	"fmt"
	"os"

	"github.com/haru-256/gce-commands/pkg/config"
	"github.com/haru-256/gce-commands/pkg/gce"
	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule-policy <vm_name> <policy_name>",
	Short: "Set schedule-policy",
	Long: `Set schedule-policy for the application.

Example:
  gce-commands set schedule-policy sandbox stop`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		policyName := args[1]
		if policyName == "" || vmName == "" {
			log.Logger.Error("schedule-policy and vm_name are required")
			os.Exit(1)
		}
		cnfPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
		// parse config
		cnf, err := config.ParseConfig(cnfPath)
		if err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
		log.Logger.Debug(fmt.Sprintf("Config: %+v", cnf))
		// filter VM by name
		vm := cnf.GetVMByName(vmName)
		// Implement your logic here
		if unset {
			log.Logger.Debug("Unset schedule-policy")
			if err = gce.UnsetSchedulePolicy(vm, policyName); err != nil {
				log.Logger.Fatal(err)
				os.Exit(1)
			}
		} else {
			log.Logger.Debug("Set schedule-policy")
			if err = gce.SetSchedulePolicy(vm, policyName); err != nil {
				log.Logger.Fatal(err)
				os.Exit(1)
			}
		}
	},
}

var unset bool

func init() {
	SetCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().BoolVarP(&unset, "un", "u", false, "Unset schedule-policy")
}
