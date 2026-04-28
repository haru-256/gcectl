package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"golang.org/x/sync/errgroup"
)

const maxConcurrentVMLookups = 10

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
// which should not be in the presentation layer. VM lookups are best-effort:
// successful lookups are returned, while failed lookups are collected into the
// returned error so the caller can still render partial results.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - configuredVMs: VMs loaded from config to query from the repository
//
// Returns:
//   - []VMListItem: Successfully retrieved VMs with calculated uptime strings
//   - error: Joined error for failed VM lookups, or nil if all lookups succeed
//
// Example:
//
//	useCase := NewListVMsUseCase(repo)
//	items, err := useCase.Execute(ctx, configuredVMs)
//	if err != nil {
//	    fmt.Fprintf(os.Stderr, "some VMs could not be listed: %v\n", err)
//	}
//	for _, item := range items {
//	    fmt.Printf("%s: %s\n", item.VM.Name, item.Uptime)
//	}
func (u *ListVMsUseCase) Execute(ctx context.Context, configuredVMs []*model.VM) ([]VMListItem, error) {
	now := time.Now()
	items := make([]VMListItem, len(configuredVMs))
	errs := make([]error, 0)
	var mu sync.Mutex

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(maxConcurrentVMLookups)

	for i, configuredVM := range configuredVMs {
		i, configuredVM := i, configuredVM
		eg.Go(func() error {
			vm, err := u.repo.FindByName(ctx, configuredVM)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("VM %s (project=%s, zone=%s): failed to find: %w", configuredVM.Name, configuredVM.Project, configuredVM.Zone, err))
				mu.Unlock()
				return nil
			}
			if vm == nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("VM %s (project=%s, zone=%s): not found", configuredVM.Name, configuredVM.Project, configuredVM.Zone))
				mu.Unlock()
				return nil
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

	successfulItems := make([]VMListItem, 0, len(items))
	for _, item := range items {
		if item.VM != nil {
			successfulItems = append(successfulItems, item)
		}
	}

	return successfulItems, errors.Join(errs...)
}
