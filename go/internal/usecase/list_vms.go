package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"golang.org/x/sync/errgroup"
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

// Execute retrieves the configured VMs and calculates their uptime strings.
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
func (u *ListVMsUseCase) Execute(ctx context.Context, configuredVMs []*model.VM) ([]VMListItem, error) {
	now := time.Now()
	items := make([]VMListItem, len(configuredVMs))
	eg, ctx := errgroup.WithContext(ctx)

	for i, configuredVM := range configuredVMs {
		i, configuredVM := i, configuredVM
		eg.Go(func() error {
			vm, err := u.repo.FindByName(ctx, configuredVM)
			if err != nil {
				return fmt.Errorf("failed to find VM %s: %w", configuredVM.Name, err)
			}
			if vm == nil {
				return fmt.Errorf("VM %s: not found", configuredVM.Name)
			}

			items[i] = VMListItem{
				VM:     vm,
				Uptime: calculateUptimeString(vm, now),
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return items, nil
}
