package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)

// offCmd represents the off command
var offCmd = &cobra.Command{
	Use:   "off <vm_name>...",
	Short: "Turn off one or more instances",
	Long: `Turn off one or more instances

Example:
  gcectl off <vm_name>
  gcectl off <vm_name1> <vm_name2> <vm_name3>`,
	Args: cobra.MinimumNArgs(1),
	Run:  offRun,
}

func offRun(cmd *cobra.Command, args []string) {
	vmNames := args
	infraLog.DefaultLogger.Debugf("Turning off the instances %s", strings.Join(vmNames, ", "))

	session, ctx, err := cli.NewSession(cmd, CnfPath)
	if err != nil {
		presenter.NewConsolePresenter().Error(err.Error())
		os.Exit(1)
	}
	defer session.Close()

	vms, err := session.Config.ResolveVMs(vmNames)
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

	stopVMUseCase := usecase.NewStopVMUseCase(session.VMRepository, infraLog.DefaultLogger)

	err = session.Console.ExecuteWithProgress(
		ctx,
		fmt.Sprintf("Stopping VMs %s", strings.Join(vmNames, ", ")),
		func(ctx context.Context) error {
			return stopVMUseCase.Execute(ctx, vms)
		},
	)
	if err != nil {
		session.Console.Error(fmt.Sprintf("Failed to turn off the instance(s): %v", err))
		session.Close()
		os.Exit(1)
	}

	session.Console.Success(fmt.Sprintf("Turned off the instances: %v", strings.Join(vmNames, ", ")))
}

func init() {
	rootCmd.AddCommand(offCmd)
}
