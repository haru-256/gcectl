/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/haru-256/gcectl/pkg/config"
	"github.com/haru-256/gcectl/pkg/gce"
	"github.com/haru-256/gcectl/pkg/log"
	"github.com/haru-256/gcectl/pkg/utils"
	"github.com/spf13/cobra"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe <vm_name>",
	Short: "Describe the instance",
	Long: `Describe the instance.

Example:
  gcectl describe <vm_name>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		log.Logger.Debugf("Describe instance %s", vmName)
		if vmName == "" {
			utils.ErrorReport("VM name is required")
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
		// Describe the instance
		// Update VMs info, such as status and schedule policy
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err = gce.UpdateInstancesInfo(ctx, []*config.VM{vm}); err != nil {
			utils.ErrorReport(fmt.Sprintf("Failed to update VMs info: %v\n", err))
			os.Exit(1)
		}

		prefixStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff79c6"))

		listItemsHeader := []string{
			"Name",
			"Project",
			"Zone",
			"MachineType",
			"Status",
			"SchedulePolicy",
		}
		itemPaddings := getItemPaddings(listItemsHeader)
		l := list.New(
			fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[0]), itemPaddings[0], vm.Name),
			fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[1]), itemPaddings[1], vm.Project),
			fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[2]), itemPaddings[2], vm.Zone),
			fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[3]), itemPaddings[3], vm.MachineType),
			fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[4]), itemPaddings[4], vm.Status),
			fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[5]), itemPaddings[5], vm.SchedulePolicy),
		).Enumerator(list.Bullet).EnumeratorStyle(lipgloss.NewStyle().Padding(0, 1))
		fmt.Println(l)
	},
}

// getItemPaddings returns paddings for each item to align the items in the list
// For example, if the listItemsHeader is ["Name", "Project", "Zone"], the return value is ["  ", " ", ""]
func getItemPaddings(listItemsHeader []string) []string {
	paddingNum := make([]int, len(listItemsHeader))
	// count max length of each item
	maxLen := 0
	for _, itemHeader := range listItemsHeader {
		if len(itemHeader) > maxLen {
			maxLen = len(itemHeader)
		}
	}
	extraPaddingNum := 1
	// calc padding for each item
	for i, itemHeader := range listItemsHeader {
		paddingNum[i] = maxLen - len(itemHeader) + extraPaddingNum
	}

	paddingsStr := make([]string, len(paddingNum))
	for i, padding := range paddingNum {
		paddingsStr[i] = ""
		if padding > 0 {
			for j := 0; j < padding; j++ {
				paddingsStr[i] += " "
			}
		}
	}
	return paddingsStr
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
