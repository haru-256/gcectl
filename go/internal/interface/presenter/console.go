package presenter

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/haru-256/gcectl/internal/domain/model"
)

var (
	purple = lipgloss.Color("99")
	gray   = lipgloss.Color("#fbfcfc ")

	headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center).Padding(0, 1)
	baseRowStyle = lipgloss.NewStyle().Padding(0, 1).Foreground(gray)
	prefixStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff79c6"))
)

// ConsolePresenter handles console output with styled messages.
// It provides methods for rendering success/error messages and VM information
// using the lipgloss library for terminal styling.
type ConsolePresenter struct {
	errorStyle   lipgloss.Style
	successStyle lipgloss.Style
}

// NewConsolePresenter creates and returns a new ConsolePresenter instance.
//
// The presenter is initialized with predefined styles:
//   - Error messages: red, bold
//   - Success messages: green, bold
//
// Returns:
//   - *ConsolePresenter: A new presenter ready for rendering output
func NewConsolePresenter() *ConsolePresenter {
	return &ConsolePresenter{
		errorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Bold(true),
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true),
	}
}

// Success prints a success message to the console with green styling.
//
// The message is prefixed with "[SUCCESS] |" and rendered in bold green color.
//
// Parameters:
//   - msg: The success message to display
//
// Example:
//
//	presenter.Success("VM started successfully")
//	// Output: [SUCCESS] | VM started successfully (in green)
func (p *ConsolePresenter) Success(msg string) {
	fmt.Println(p.successStyle.Render("[SUCCESS] | ") + msg)
}

// Error prints an error message to the console with red styling.
//
// The message is prefixed with "[ERROR] |" and rendered in bold red color.
//
// Parameters:
//   - msg: The error message to display
//
// Example:
//
//	presenter.Error("Failed to start VM: not found")
//	// Output: [ERROR] | Failed to start VM: not found (in red)
func (p *ConsolePresenter) Error(msg string) {
	fmt.Println(p.errorStyle.Render("[ERROR] | ") + msg)
}

// RenderVMList renders a list of VMs in a formatted table.
//
// The table includes the following columns:
//   - Name: VM instance name
//   - Project: GCP project ID
//   - Zone: GCP zone
//   - Machine-Type: VM machine type
//   - Status: Current status with emoji indicator (🟢 for RUNNING, 🔴 for STOPPED/TERMINATED)
//   - Schedule: Attached schedule policy name (if any)
//   - Uptime: How long the VM has been running (for RUNNING VMs only)
//
// Parameters:
//   - vms: Slice of VM instances to render
//
// Example output:
//
//	┌──────────┬────────────┬──────────────┬──────────────┬─────────────┬──────────┬─────────┐
//	│   Name   │  Project   │     Zone     │ Machine-Type │   Status    │ Schedule │ Uptime  │
//	├──────────┼────────────┼──────────────┼──────────────┼─────────────┼──────────┼─────────┤
//	│ my-vm    │ my-project │ us-central1-a│ e2-medium    │ 🟢 RUNNING  │ policy-1 │ 2h30m   │
//	└──────────┴────────────┴──────────────┴──────────────┴─────────────┴──────────┴─────────┘
func (p *ConsolePresenter) RenderVMList(vms []*model.VM) {
	var rows [][]string
	now := time.Now()

	for _, vm := range vms {
		uptime := "N/A"
		if vm.LastStartTime != nil && vm.Status.String() == "RUNNING" {
			duration := now.Sub(*vm.LastStartTime)
			uptime = duration.String()
		}

		statusEmoji := "⚪"
		switch vm.Status.String() {
		case "RUNNING":
			statusEmoji = "🟢"
		case "STOPPED", "TERMINATED":
			statusEmoji = "🔴"
		}

		rows = append(rows, []string{
			vm.Name,
			vm.Project,
			vm.Zone,
			vm.MachineType,
			statusEmoji + " " + vm.Status.String(),
			vm.SchedulePolicy,
			uptime,
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		Headers("Name", "Project", "Zone", "Machine-Type", "Status", "Schedule", "Uptime").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return headerStyle
			default:
				return baseRowStyle.Align(lipgloss.Left)
			}
		})

	fmt.Println(t)
}

// RenderVMDetail renders detailed information about a single VM in a list format.
//
// The output includes:
//   - Name: VM instance name
//   - Project: GCP project ID
//   - Zone: GCP zone
//   - MachineType: VM machine type
//   - Status: Current operational status
//   - SchedulePolicy: Attached schedule policy (if any)
//
// All fields are aligned for readability with bullet points.
//
// Parameters:
//   - vm: The VM instance to render details for
//
// Example output:
//
//   - Name          : my-vm
//   - Project       : my-project
//   - Zone          : us-central1-a
//   - MachineType   : e2-medium
//   - Status        : RUNNING
//   - SchedulePolicy: my-schedule-policy
func (p *ConsolePresenter) RenderVMDetail(vm *model.VM) {
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
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[4]), itemPaddings[4], vm.Status.String()),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[5]), itemPaddings[5], vm.SchedulePolicy),
	).Enumerator(list.Bullet).EnumeratorStyle(lipgloss.NewStyle().Padding(0, 1))

	fmt.Println(l)
}

// getItemPaddings calculates padding strings for list items to ensure alignment.
//
// This helper function determines how much padding each item needs based on
// the longest item header in the list, ensuring all colons align vertically.
//
// Parameters:
//   - listItemsHeader: Slice of header strings (e.g., ["Name", "Project", "Zone"])
//
// Returns:
//   - []string: Slice of padding strings, one for each header
//
// Example:
//
//	headers := []string{"Name", "Project", "Zone"}
//	paddings := getItemPaddings(headers)
//	// Returns: ["   ", "  ", "     "] to align all items to "Project" length
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
