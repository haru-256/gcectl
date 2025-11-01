package usecase

import (
	"context"
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
)

// UnsetSchedulePolicyUseCase handles the business logic for removing a schedule policy
type UnsetSchedulePolicyUseCase struct {
	vmRepo repository.VMRepository
	logger log.Logger
}

// NewUnsetSchedulePolicyUseCase creates a new instance of UnsetSchedulePolicyUseCase
func NewUnsetSchedulePolicyUseCase(vmRepo repository.VMRepository, logger log.Logger) *UnsetSchedulePolicyUseCase {
	return &UnsetSchedulePolicyUseCase{vmRepo: vmRepo, logger: logger}
}

// Execute removes a schedule policy from a VM.
//
// This method performs the following steps:
// 1. Retrieves the VM instance from the repository
// 2. Executes the schedule policy removal operation
//
// Removing a schedule policy stops the automatic start/stop behavior controlled by that policy.
//
// Parameters:
//   - ctx: The context for the operation (used for cancellation and timeout)
//   - project: The GCP project ID
//   - zone: The GCP zone
//   - name: The VM instance name
//   - policyName: The name of the schedule policy to remove
//
// Returns:
//   - error: nil on success, otherwise an error describing what went wrong
//
// Error conditions:
//   - VM not found: when the VM does not exist in the specified project/zone
//   - Unset operation failed: when the GCP API call to remove the schedule policy fails
//
// Example:
//
//	usecase := NewUnsetSchedulePolicyUseCase(vmRepo)
//	err := usecase.Execute(ctx, "my-project", "us-central1-a", "my-vm", "my-schedule-policy")
//	if err != nil {
//	    log.Fatalf("Failed to unset schedule policy: %v", err)
//	}
func (uc *UnsetSchedulePolicyUseCase) Execute(ctx context.Context, project, zone, name, policyName string) error {
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

	// 2. スケジュールポリシー削除実行
	if unsetErr := uc.vmRepo.UnsetSchedulePolicy(ctx, foundVM, policyName); unsetErr != nil {
		return fmt.Errorf("failed to unset schedule policy: %w", unsetErr)
	}

	uc.logger.Infof("✓ Successfully unset schedule policy %s for VM %s", policyName, foundVM.Name)
	return nil
}
