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
	"github.com/spf13/cobra"
)

// onCmd represents the on command
var onCmd = &cobra.Command{
	Use:   "on <vm_name>",
	Short: "Turn on the instance",
	Long: `Turn on the instance

Example:
  gce on <vm_name>`,
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
			log.Logger.Fatal(err)
			os.Exit(1)
		}
		log.Logger.Debug(fmt.Sprintf("Config: %+v", cnf))

		// filter VM by name
		vm := cnf.GetVMByName(vmName)

		// Turn on the instance
		if err = gce.OnVM(vm); err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(onCmd)
}
