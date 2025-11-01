package presenter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/haru-256/gcectl/internal/domain/model"
	"golang.org/x/sync/errgroup"
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

// ProgressStart prints a progress message without a newline.
// This is typically called at the start of long-running operations.
//
// The message is displayed as-is, allowing the Progress() method to add
// dots on the same line, followed by ProgressDone() to complete the line.
//
// Parameters:
//   - msg: The progress message to display (e.g., "Starting VM my-vm")
//
// Example:
//
//	presenter.ProgressStart("Starting VM my-vm")
//	// ... operation in progress, Progress() called multiple times ...
//	presenter.ProgressDone()
//	// Output: Starting VM my-vm...
func (p *ConsolePresenter) ProgressStart(msg string) {
	fmt.Print(msg)
}

// Progress prints a progress indicator (dot) without a newline.
// This is typically called periodically during long-running operations.
//
// Example:
//
//	// During operation: . . . . .
//	presenter.Progress()
func (p *ConsolePresenter) Progress() {
	fmt.Print(".")
}

// ProgressDone prints a newline to complete a progress indicator line.
// This should be called after a series of Progress() calls.
//
// Example:
//
//	presenter.Progress() // prints "."
//	presenter.Progress() // prints "."
//	presenter.ProgressDone() // prints newline
func (p *ConsolePresenter) ProgressDone() {
	fmt.Println()
}

// VMListItem represents a VM instance for list view display.
// This type is used to decouple the presenter layer from domain models,
// allowing the presentation logic to receive pre-formatted data.
//
//nolint:govet // Field order optimized for readability over memory alignment
type VMListItem struct {
	Name           string
	Project        string
	Zone           string
	MachineType    string
	Status         model.Status
	SchedulePolicy string
	Uptime         string // Pre-calculated uptime string (e.g., "7d12h45m", "2h30m", "5m30s", "45s", "N/A")
}

// VMDetail is an alias for VMListItem since they have identical structure.
// This improves code clarity by using different names for different contexts,
// while avoiding duplication of the type definition.
type VMDetail = VMListItem

// getStatusEmoji returns an emoji representation of a VM status.
//
// This helper function centralizes the status-to-emoji mapping logic,
// ensuring consistent status indicators across different display formats.
//
// Parameters:
//   - status: The VM status to get an emoji for
//
// Returns:
//   - string: Emoji representing the status (🟢 for RUNNING, 🔴 for STOPPED/TERMINATED, ⚪ for others)
//
// Example:
//
//	emoji := getStatusEmoji(model.StatusRunning)
//	// Returns: "🟢"
func getStatusEmoji(status model.Status) string {
	switch status.String() {
	case "RUNNING":
		return "🟢"
	case "STOPPED", "TERMINATED":
		return "🔴"
	default:
		return "⚪"
	}
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
//   - Uptime: How long the VM has been running
//     Format: "7d12h45m" (days), "2h30m" (hours), "5m30s" (minutes), "45s" (seconds), "N/A" (stopped)
//
// The uptime string is expected to be pre-calculated by the use case layer,
// keeping business logic out of the presentation layer.
//
// Parameters:
//   - items: Slice of VMListItem with VMs and their pre-calculated uptime strings
//
// Example output:
//
//	┌──────────┬────────────┬──────────────┬──────────────┬─────────────┬──────────┬─────────┐
//	│   Name   │  Project   │     Zone     │ Machine-Type │   Status    │ Schedule │ Uptime  │
//	├──────────┼────────────┼──────────────┼──────────────┼─────────────┼──────────┼─────────┤
//	│ my-vm    │ my-project │ us-central1-a│ e2-medium    │ 🟢 RUNNING  │ policy-1 │ 2h30m   │
//	│ dev-vm   │ my-project │ us-west1-a   │ n1-standard-1│ 🟢 RUNNING  │          │ 7d12h45m│
//	└──────────┴────────────┴──────────────┴──────────────┴─────────────┴──────────┴─────────┘
func (p *ConsolePresenter) RenderVMList(items []VMListItem) {
	var rows [][]string

	for _, item := range items {
		statusEmoji := getStatusEmoji(item.Status)

		rows = append(rows, []string{
			item.Name,
			item.Project,
			item.Zone,
			item.MachineType,
			statusEmoji + " " + item.Status.String(),
			item.SchedulePolicy,
			item.Uptime,
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
//   - Uptime: How long the VM has been running
//     Format: "7d12h45m" (days), "2h30m" (hours), "5m30s" (minutes), "45s" (seconds), "N/A" (stopped)
//
// All fields are aligned for readability with bullet points.
//
// Parameters:
//   - detail: The VM detail information with pre-calculated uptime string
//
// Example output:
//
//   - Name          : my-vm
//   - Project       : my-project
//   - Zone          : us-central1-a
//   - MachineType   : e2-medium
//   - Status        : RUNNING
//   - SchedulePolicy: my-schedule-policy
//   - Uptime        : 2h30m
func (p *ConsolePresenter) RenderVMDetail(detail VMDetail) {
	listItemsHeader := []string{
		"Name",
		"Project",
		"Zone",
		"MachineType",
		"Status",
		"SchedulePolicy",
		"Uptime",
	}
	itemPaddings := getItemPaddings(listItemsHeader)

	l := list.New(
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[0]), itemPaddings[0], detail.Name),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[1]), itemPaddings[1], detail.Project),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[2]), itemPaddings[2], detail.Zone),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[3]), itemPaddings[3], detail.MachineType),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[4]), itemPaddings[4], detail.Status.String()),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[5]), itemPaddings[5], detail.SchedulePolicy),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render(listItemsHeader[6]), itemPaddings[6], detail.Uptime),
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
			paddingsStr[i] = strings.Repeat(" ", padding)
		}
	}
	return paddingsStr
}

// ExecuteWithProgress executes a function with progress indication.
//
// This function displays a progress message, executes the provided function
// in a goroutine, and shows progress dots every second until completion.
// It properly handles context cancellation and ensures clean shutdown.
//
// Parameters:
//   - ctx: Context for cancellation control
//   - message: Initial progress message (e.g., "Starting VMs")
//   - fn: The function to execute (receives context and returns error)
//
// Returns:
//   - error: Error from the executed function, or nil on success
//
// Example:
//
//	err := console.ExecuteWithProgress(
//	    ctx,
//	    "Starting VMs vm-1, vm-2",
//	    func(ctx context.Context) error {
//	        return startVMUseCase.Execute(ctx, vms)
//	    },
//	)
func (p *ConsolePresenter) ExecuteWithProgress(ctx context.Context, message string, fn func(context.Context) error) error {
	p.ProgressStart(message)
	defer p.ProgressDone()

	eg, ctx := errgroup.WithContext(ctx)
	doneCh := make(chan struct{})

	// Execute the function
	eg.Go(func() error {
		defer close(doneCh)
		if err := fn(ctx); err != nil {
			return err
		}
		return nil
	})

	// Display progress dots every second
	eg.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-doneCh:
				return nil
			case <-ticker.C:
				p.Progress()
			}
		}
	})

	return eg.Wait()
}
