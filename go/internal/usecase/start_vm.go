package usecase

import (
	"context"
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"golang.org/x/sync/errgroup"
)

// StartVMUseCase handles the business logic for starting a VM
type StartVMUseCase struct {
	vmRepo repository.VMRepository
}

// NewStartVMUseCase creates a new instance of StartVMUseCase
func NewStartVMUseCase(vmRepo repository.VMRepository) *StartVMUseCase {
	return &StartVMUseCase{vmRepo: vmRepo}
}

// Execute starts multiple VM instances in parallel.
// All VMs are processed concurrently. If any VM fails, the entire operation is canceled (fail-fast).
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - vms: VMs to start (must contain Project, Zone, and Name)
//
// Returns:
//   - error: nil on success, or error with VM name on failure
func (uc *StartVMUseCase) Execute(ctx context.Context, vms []*model.VM) error {
	// TOCTOU問題に対応するため、1つのgoroutineのなかでCheckとUseを実行する
	eg, ctx := errgroup.WithContext(ctx)
	for _, vm := range vms {
		vm := vm // capture range variable
		eg.Go(func() error {
			// 1. VMが存在するか確認
			foundVM, err := uc.vmRepo.FindByName(ctx, vm)
			if err != nil {
				return fmt.Errorf("VM %s: failed to find: %w", vm.Name, err)
			}
			if foundVM == nil {
				return fmt.Errorf("VM %s: not found", vm.Name)
			}

			// 2. ビジネスルールチェック
			if !foundVM.CanStart() {
				return fmt.Errorf("VM %s: cannot be started (current status: %s)",
					foundVM.Name, foundVM.Status)
			}

			// 3. 起動実行
			if startErr := uc.vmRepo.Start(ctx, foundVM); startErr != nil {
				return fmt.Errorf("VM %s: failed to start: %w", foundVM.Name, startErr)
			}

			return nil
		})
	}

	return eg.Wait()
}
