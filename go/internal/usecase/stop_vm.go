package usecase

import (
	"context"
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
	"golang.org/x/sync/errgroup"
)

// StopVMUseCase handles the business logic for stopping a VM
type StopVMUseCase struct {
	vmRepo repository.VMRepository
	logger log.Logger
}

// NewStopVMUseCase creates a new instance of StopVMUseCase
func NewStopVMUseCase(vmRepo repository.VMRepository, logger log.Logger) *StopVMUseCase {
	return &StopVMUseCase{vmRepo: vmRepo, logger: logger}
}

// Execute stops multiple VM instances in parallel after validating each can be stopped.
//
// Parameters:
//   - ctx: The context for the operation
//   - vms: The VM instances to stop
//
// Returns:
//   - error: nil on success, otherwise an error describing what went wrong
func (uc *StopVMUseCase) Execute(ctx context.Context, vms []*model.VM) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, vm := range vms {
		vm := vm
		eg.Go(func() error {
			// 1. VMを取得して存在確認
			foundVM, err := uc.vmRepo.FindByName(ctx, vm)
			if err != nil {
				return fmt.Errorf("VM %s: failed to find: %w", vm.Name, err)
			}

			if foundVM == nil {
				return fmt.Errorf("VM %s: not found", vm.Name)
			}

			// 2. ビジネスルールチェック
			if !foundVM.CanStop() {
				return fmt.Errorf("VM %s: cannot be stopped (current status: %s)", foundVM.Name, foundVM.Status)
			}

			// 3. 停止実行
			if stopErr := uc.vmRepo.Stop(ctx, foundVM); stopErr != nil {
				return fmt.Errorf("VM %s: failed to stop: %w", foundVM.Name, stopErr)
			}

			uc.logger.Infof("✓ Successfully stopped VM %s", foundVM.Name)
			return nil
		})
	}

	return eg.Wait()
}
