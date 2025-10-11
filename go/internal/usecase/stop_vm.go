package usecase

import (
	"context"
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/repository"
)

// StopVMUseCase handles the business logic for stopping a VM
type StopVMUseCase struct {
	vmRepo repository.VMRepository
}

// NewStopVMUseCase creates a new instance of StopVMUseCase
func NewStopVMUseCase(vmRepo repository.VMRepository) *StopVMUseCase {
	return &StopVMUseCase{vmRepo: vmRepo}
}

// Execute stops a VM instance after validating it can be stopped.
//
// This method performs the following steps:
// 1. Retrieves the VM instance from the repository
// 2. Validates that the VM can be stopped (business rule check via CanStop)
// 3. Executes the stop operation
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
//   - VM cannot be stopped: when the VM is not in a stoppable state (e.g., already TERMINATED)
//   - Stop operation failed: when the GCP API call to stop the VM fails
//
// Example:
//
//	usecase := NewStopVMUseCase(vmRepo)
//	err := usecase.Execute(ctx, "my-project", "us-central1-a", "my-vm")
//	if err != nil {
//	    log.Fatalf("Failed to stop VM: %v", err)
//	}
func (uc *StopVMUseCase) Execute(ctx context.Context, project, zone, name string) error {
	// 1. VMを取得
	vm, err := uc.vmRepo.FindByName(ctx, project, zone, name)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	// 2. ビジネスルールチェック
	if !vm.CanStop() {
		return fmt.Errorf("VM %s cannot be stopped (current status: %s)", vm.Name, vm.Status)
	}

	// 3. 停止実行
	if stopErr := uc.vmRepo.Stop(ctx, vm); stopErr != nil {
		return fmt.Errorf("failed to stop VM: %w", stopErr)
	}

	return nil
}
