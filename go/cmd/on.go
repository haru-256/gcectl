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

// onCmd represents the on command
var onCmd = &cobra.Command{
	Use:   "on <vm_name...>",
	Short: "Turn on the instances",
	Long: `Turn on the instances

Example:
  gcectl on <vm_name>
  gcectl on <vm_name1> <vm_name2> <vm_name3>`,
	Args: cobra.MinimumNArgs(1),
	Run:  onRun,
}

func onRun(cmd *cobra.Command, args []string) {
	console := presenter.NewConsolePresenter()
	vmNames := args
	infraLog.DefaultLogger.Debugf("Turning on the instances %s", strings.Join(vmNames, ", "))

	session, ctx, err := cli.NewSession(cmd, CnfPath)
	if err != nil {
		console.Error(err.Error())
		os.Exit(1)
	}
	defer session.Close()

	vms, err := session.Config.ResolveVMs(vmNames)
	if err != nil {
		console.Error(err.Error())
		session.Close()
		os.Exit(1)
	}

	err = session.OpenVMRepository(ctx)
	if err != nil {
		console.Error(err.Error())
		session.Close()
		os.Exit(1)
	}

	startVMUseCase := usecase.NewStartVMUseCase(session.VMRepository, infraLog.DefaultLogger)

	err = console.ExecuteWithProgress(
		ctx,
		fmt.Sprintf("Starting VMs %s", strings.Join(vmNames, ", ")),
		func(ctx context.Context) error {
			return startVMUseCase.Execute(ctx, vms)
		},
	)
	if err != nil {
		console.Error(fmt.Sprintf("Failed to turn on the instances: %v", err))
		session.Close()
		os.Exit(1)
	}

	console.Success(fmt.Sprintf("Turned on the instances: %v", strings.Join(vmNames, ", ")))
}

func init() {
	rootCmd.AddCommand(onCmd)
}
