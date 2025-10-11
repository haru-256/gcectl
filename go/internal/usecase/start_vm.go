package usecase

import (
	"context"
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/repository"
)

// StartVMUseCase handles the business logic for starting a VM
type StartVMUseCase struct {
	vmRepo repository.VMRepository
}

// NewStartVMUseCase creates a new instance of StartVMUseCase
func NewStartVMUseCase(vmRepo repository.VMRepository) *StartVMUseCase {
	return &StartVMUseCase{vmRepo: vmRepo}
}

// Execute starts a VM instance after validating it can be started.
//
// This method performs the following steps:
// 1. Retrieves the VM instance from the repository
// 2. Validates that the VM can be started (business rule check via CanStart)
// 3. Executes the start operation
//
// Parameters:
//   - ctx: The context for the operation (used for cancellation and timeout)
//   - project: The GCP project ID
//   - zone: The GCP zone
//   - name: The VM instance name
//
// Returns:
//   - error: nil on success, otherwise an error describing what went wrong
//
// Error conditions:
//   - VM not found: when the VM does not exist in the specified project/zone
//   - VM cannot be started: when the VM is not in a startable state (e.g., already RUNNING)
//   - Start operation failed: when the GCP API call to start the VM fails
//
// Example:
//
//	usecase := NewStartVMUseCase(vmRepo)
//	err := usecase.Execute(ctx, "my-project", "us-central1-a", "my-vm")
//	if err != nil {
//	    log.Fatalf("Failed to start VM: %v", err)
//	}
func (uc *StartVMUseCase) Execute(ctx context.Context, project, zone, name string) error {
	// 1. VMを取得
	vm, err := uc.vmRepo.FindByName(ctx, project, zone, name)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	// 2. ビジネスルールチェック
	if !vm.CanStart() {
		return fmt.Errorf("VM %s cannot be started (current status: %s)", vm.Name, vm.Status)
	}

	// 3. 起動実行
	if startErr := uc.vmRepo.Start(ctx, vm); startErr != nil {
		return fmt.Errorf("failed to start VM: %w", startErr)
	}

	return nil
}
