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
type ConsolePresenter struct {
	errorStyle   lipgloss.Style
	successStyle lipgloss.Style
}

// NewConsolePresenter creates a new ConsolePresenter instance.
//
// Returns:
//   - *ConsolePresenter: A new presenter with predefined styles
func NewConsolePresenter() *ConsolePresenter {
	return &ConsolePresenter{
		errorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Bold(true),
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true),
	}
}

// Success prints a success message with green styling.
//
// Parameters:
//   - msg: The success message to display
func (p *ConsolePresenter) Success(msg string) {
	fmt.Println(p.successStyle.Render("[SUCCESS] | ") + msg)
}

// Error prints an error message with red styling.
//
// Parameters:
//   - msg: The error message to display
func (p *ConsolePresenter) Error(msg string) {
	fmt.Println(p.errorStyle.Render("[ERROR] | ") + msg)
}

// ProgressStart prints a progress message without a newline.
//
// Parameters:
//   - msg: The progress message to display
func (p *ConsolePresenter) ProgressStart(msg string) {
	fmt.Print(msg)
}

// Progress prints a dot (.) without a newline for progress indication.
func (p *ConsolePresenter) Progress() {
	fmt.Print(".")
}

// ProgressDone prints a newline to complete a progress line.
func (p *ConsolePresenter) ProgressDone() {
	fmt.Println()
}

// VMListItem represents a VM instance for display.
//
//nolint:govet // Field order optimized for readability
type VMListItem struct {
	Name           string
	Project        string
	Zone           string
	MachineType    string
	Status         model.Status
	SchedulePolicy string
	Uptime         string // Pre-calculated uptime (e.g., "7d12h45m", "2h30m", "5m30s", "N/A")
}

// VMDetail is an alias for VMListItem for code clarity.
type VMDetail = VMListItem

// getStatusEmoji returns an emoji for the given VM status.
//
// Parameters:
//   - status: The VM status
//
// Returns:
//   - string: 🟢 for RUNNING, 🔴 for STOPPED/TERMINATED, ⚪ for others
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

// RenderVMList renders VMs in a formatted table.
//
// Parameters:
//   - items: VMs to display with pre-calculated uptime strings
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

// RenderVMDetail renders detailed VM information in a list format.
//
// Parameters:
//   - detail: VM details with pre-calculated uptime string
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

// RenderVersion renders version information in a list format.
//
// Parameters:
//   - version: Version string
//   - commit: Git commit hash
//   - date: Build date
func (p *ConsolePresenter) RenderVersion(version, commit, date string) {
	itemPaddings := getItemPaddings([]string{
		"Version",
		"Git Commit",
		"Build Date",
	})
	l := list.New(
		fmt.Sprintf("%s%s: %s", prefixStyle.Render("Version"), itemPaddings[0], version),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render("Git Commit"), itemPaddings[1], commit),
		fmt.Sprintf("%s%s: %s", prefixStyle.Render("Build Date"), itemPaddings[2], date),
	).Enumerator(list.Bullet).EnumeratorStyle(lipgloss.NewStyle().Padding(0, 1))

	fmt.Println(l)
}

// getItemPaddings calculates padding strings for list items to ensure alignment.
//
// Parameters:
//   - listItemsHeader: Slice of header strings
//
// Returns:
//   - []string: Slice of padding strings, one for each header
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
// Displays a progress message, executes the provided function in a goroutine,
// and shows progress dots every second until completion.
//
// Parameters:
//   - ctx: Context for cancellation control
//   - message: Initial progress message (e.g., "Starting VMs")
//   - fn: The function to execute (receives context and returns error)
//
// Returns:
//   - error: Error from the executed function, or nil on success
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
