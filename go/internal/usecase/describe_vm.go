package usecase

import (
	"context"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
)

// DescribeVM retrieves detailed information about a specific VM and returns it with a calculated uptime string.
//
// This use case encapsulates the business logic of fetching a VM and calculating its uptime,
// keeping this logic out of the presentation layer. The uptime is returned as a formatted string
// ready for display.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - repo: VM repository interface for data access
//   - project: GCP project ID
//   - zone: GCP zone
//   - name: VM instance name
//
// Returns:
//   - *model.VM: The VM instance with current status
//   - string: Formatted uptime string (e.g., "2h30m" for running VMs, "N/A" for stopped VMs)
//   - error: Error if VM retrieval fails
//
// Example:
//
//	vm, uptime, err := DescribeVM(ctx, repo, "my-project", "us-central1-a", "my-vm")
//	if err != nil {
//	    return err
//	}
//	// vm: &model.VM{Name: "my-vm", Status: model.StatusRunning, ...}
//	// uptime: "2h30m15s"
func DescribeVM(ctx context.Context, repo repository.VMRepository, project, zone, name string) (*model.VM, string, error) {
	vm, err := repo.FindByName(ctx, project, zone, name)
	if err != nil {
		return nil, "", err
	}

	// Calculate uptime using shared logic
	uptimeStr := calculateUptimeString(vm, time.Now())

	return vm, uptimeStr, nil
}
