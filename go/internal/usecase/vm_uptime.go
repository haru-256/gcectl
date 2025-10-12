package usecase

import (
	"fmt"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
)

// calculateUptimeString calculates the uptime string for a VM.
//
// This helper function encapsulates the common logic for calculating uptime
// across different use cases. It:
//   - Returns "N/A" if the VM is not running
//   - Calls the VM's Uptime() method to get the duration
//   - Formats the duration in a human-readable format
//   - Returns "N/A" if uptime calculation fails
//
// Format rules:
//   - Minimum unit: seconds (no milliseconds)
//   - Days and above: shows days, hours, and minutes (e.g., "2d5h30m")
//   - Hours and above: shows hours and minutes only (e.g., "2h30m")
//   - Minutes only: shows minutes and seconds (e.g., "5m30s")
//   - Seconds only: shows seconds (e.g., "45s")
//
// This keeps the uptime calculation logic centralized and consistent
// across all use cases (list, describe, etc.).
//
// Parameters:
//   - vm: The VM instance to calculate uptime for
//   - now: The current time to calculate uptime against
//
// Returns:
//   - string: Formatted uptime string (e.g., "2d5h30m", "2h30m", "5m30s", "N/A")
//
// Example:
//
//	uptimeStr := calculateUptimeString(vm, time.Now())
//	// Returns: "2d5h30m" for days, "2h30m" for hours, "5m30s" for minutes, "N/A" for stopped VMs
func calculateUptimeString(vm *model.VM, now time.Time) string {
	uptime, err := vm.Uptime(now)
	if err != nil {
		return "N/A"
	}
	return formatUptime(uptime)
}

// formatUptime formats a duration into a human-readable uptime string.
//
// Format rules:
//   - If duration >= 1 day: "Xd Yh Zm" (days, hours, and minutes)
//   - If duration >= 1 hour: "Xh Ym" (hours and minutes only)
//   - If duration >= 1 minute: "Xm Ys" (minutes and seconds)
//   - If duration < 1 minute: "Xs" (seconds only)
//   - Minimum unit is seconds (no milliseconds)
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - string: Formatted uptime string
//
// Example:
//
//	formatUptime(2*24*time.Hour + 5*time.Hour + 30*time.Minute) // "2d5h30m"
//	formatUptime(2*time.Hour + 30*time.Minute + 15*time.Second) // "2h30m"
//	formatUptime(5*time.Minute + 30*time.Second)                // "5m30s"
//	formatUptime(45*time.Second)                                // "45s"
func formatUptime(d time.Duration) string {
	// Round to seconds
	d = d.Round(time.Second)

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		// Days, hours, and minutes (no seconds)
		return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
	} else if hours > 0 {
		// Hours and minutes only (no seconds)
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else if minutes > 0 {
		// Minutes and seconds
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		// Seconds only
		return fmt.Sprintf("%ds", seconds)
	}
}
