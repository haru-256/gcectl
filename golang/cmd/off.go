/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/haru-256/gce-commands/pkg/config"
	"github.com/haru-256/gce-commands/pkg/gce"
	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/haru-256/gce-commands/pkg/utils"
	"github.com/spf13/cobra"
)

// offCmd represents the off command
var offCmd = &cobra.Command{
	Use:   "off <vm_name>",
	Short: "Turn off the instance",
	Long: `Turn off the instance

Example:
  gce off <vm_name>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		log.Logger.Debugf("Turning on the instance %s", vmName)
		if vmName == "" {
			log.Logger.Error("VM name is required")
			os.Exit(1)
		}
		// parse config
		cnf, err := config.ParseConfig(CnfPath)
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

		// Turn off the instance
		if err = gce.OffVM(vm); err != nil {
			utils.ErrorReport(fmt.Sprintf("Failed to turn off the instance: %v\n", err))
			os.Exit(1)
		}
		utils.SuccessReport(fmt.Sprintf("Turned off the instance: %v\n", vmName))
	},
}

func init() {
	rootCmd.AddCommand(offCmd)
}
