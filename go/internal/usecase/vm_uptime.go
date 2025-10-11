package usecase

import (
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
)

// calculateUptimeString calculates the uptime string for a VM.
//
// This helper function encapsulates the common logic for calculating uptime
// across different use cases. It:
//   - Returns "N/A" if the VM is not running
//   - Calls the VM's Uptime() method to get the duration
//   - Converts the duration to a string format (e.g., "2h30m15s")
//   - Returns "N/A" if uptime calculation fails
//
// This keeps the uptime calculation logic centralized and consistent
// across all use cases (list, describe, etc.).
//
// Parameters:
//   - vm: The VM instance to calculate uptime for
//   - now: The current time to calculate uptime against
//
// Returns:
//   - string: Formatted uptime string (e.g., "2h30m" for running VMs, "N/A" for stopped VMs)
//
// Example:
//
//	uptimeStr := calculateUptimeString(vm, time.Now())
//	// Returns: "2h30m15s" for a running VM, "N/A" for a stopped VM
func calculateUptimeString(vm *model.VM, now time.Time) string {
	uptime, err := vm.Uptime(now)
	if err != nil {
		return "N/A"
	}
	return uptime.String()
}
