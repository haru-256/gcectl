package usecase

import (
	"context"
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
)

// UpdateMachineTypeUseCase handles the business logic for updating VM machine type
type UpdateMachineTypeUseCase struct {
	vmRepo repository.VMRepository
	logger log.Logger
}

// NewUpdateMachineTypeUseCase creates a new instance of UpdateMachineTypeUseCase
func NewUpdateMachineTypeUseCase(vmRepo repository.VMRepository, logger log.Logger) *UpdateMachineTypeUseCase {
	return &UpdateMachineTypeUseCase{vmRepo: vmRepo, logger: logger}
}

// Execute updates the machine type of a VM after validating it is in a stopped state.
//
// This method performs the following steps:
// 1. Retrieves the VM instance from the repository
// 2. Validates that the VM is stopped (business rule: cannot change machine type of running VM)
// 3. Executes the machine type update operation
//
// Parameters:
//   - ctx: The context for the operation (used for cancellation and timeout)
//   - project: The GCP project ID
//   - zone: The GCP zone
//   - name: The VM instance name
//   - machineType: The new machine type (e.g., "e2-medium", "n1-standard-1")
//
// Returns:
//   - error: nil on success, otherwise an error describing what went wrong
//
// Error conditions:
//   - VM not found: when the VM does not exist in the specified project/zone
//   - VM is running: when the VM is not stopped (machine type can only be changed when VM is TERMINATED)
//   - Update operation failed: when the GCP API call to update the machine type fails
//
// Example:
//
//	usecase := NewUpdateMachineTypeUseCase(vmRepo)
//	err := usecase.Execute(ctx, "my-project", "us-central1-a", "my-vm", "e2-medium")
//	if err != nil {
//	    log.Fatalf("Failed to update machine type: %v", err)
//	}
func (uc *UpdateMachineTypeUseCase) Execute(ctx context.Context, project, zone, name, machineType string) error {
	// 1. VMを取得
	vm := &model.VM{
		Project: project,
		Zone:    zone,
		Name:    name,
	}
	foundVM, err := uc.vmRepo.FindByName(ctx, vm)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	// 2. ビジネスルールチェック（VMは停止状態である必要がある）
	if foundVM.CanStop() {
		return fmt.Errorf("VM %s must be stopped before changing machine type (current status: %s)", foundVM.Name, foundVM.Status)
	}

	// 3. マシンタイプ更新実行
	if updateErr := uc.vmRepo.UpdateMachineType(ctx, foundVM, machineType); updateErr != nil {
		return fmt.Errorf("failed to update machine type: %w", updateErr)
	}

	uc.logger.Infof("✓ Successfully updated machine type to %s for VM %s", machineType, foundVM.Name)
	return nil
}
