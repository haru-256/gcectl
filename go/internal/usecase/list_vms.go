package usecase

import (
	"context"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
)

// VMListItem represents a VM with its display information including uptime.
// This struct is used to pass presentation-ready data from the use case layer
// to the presenter layer, keeping business logic out of the presentation layer.
type VMListItem struct {
	VM     *model.VM
	Uptime string
}

// ListVMsUseCase handles the business logic for listing VMs with their uptime.
type ListVMsUseCase struct {
	repo repository.VMRepository
}

// NewListVMsUseCase creates a new ListVMsUseCase instance.
//
// Parameters:
//   - repo: The VM repository for data access
//
// Returns:
//   - *ListVMsUseCase: A new use case instance
func NewListVMsUseCase(repo repository.VMRepository) *ListVMsUseCase {
	return &ListVMsUseCase{
		repo: repo,
	}
}

// Execute retrieves all VMs and calculates their uptime strings.
//
// This method encapsulates the business logic of calculating uptime,
// which should not be in the presentation layer. For each VM, it:
//   - Calls the shared calculateUptimeString() function
//   - Returns "N/A" for VMs that are not running or have errors
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - []VMListItem: List of VMs with their calculated uptime strings
//   - error: Error if VM retrieval fails
//
// Example:
//
//	useCase := NewListVMsUseCase(repo)
//	items, err := useCase.Execute(ctx)
//	if err != nil {
//	    return err
//	}
//	for _, item := range items {
//	    fmt.Printf("%s: %s\n", item.VM.Name, item.Uptime)
//	}
func (u *ListVMsUseCase) Execute(ctx context.Context) ([]VMListItem, error) {
	vms, err := u.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]VMListItem, len(vms))
	for i, vm := range vms {
		items[i] = VMListItem{
			VM:     vm,
			Uptime: calculateUptimeString(vm, now),
		}
	}

	return items, nil
}
