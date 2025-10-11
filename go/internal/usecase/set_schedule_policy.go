package usecase

import (
	"context"
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/repository"
)

// SetSchedulePolicyUseCase handles the business logic for setting a schedule policy
type SetSchedulePolicyUseCase struct {
	vmRepo repository.VMRepository
}

// NewSetSchedulePolicyUseCase creates a new instance of SetSchedulePolicyUseCase
func NewSetSchedulePolicyUseCase(vmRepo repository.VMRepository) *SetSchedulePolicyUseCase {
	return &SetSchedulePolicyUseCase{vmRepo: vmRepo}
}

// Execute attaches a schedule policy to a VM.
//
// This method performs the following steps:
// 1. Retrieves the VM instance from the repository
// 2. Executes the schedule policy attachment operation
//
// A schedule policy controls when the VM should be automatically started or stopped.
//
// Parameters:
//   - ctx: The context for the operation (used for cancellation and timeout)
//   - project: The GCP project ID
//   - zone: The GCP zone
//   - name: The VM instance name
//   - policyName: The name of the schedule policy to attach
//
// Returns:
//   - error: nil on success, otherwise an error describing what went wrong
//
// Error conditions:
//   - VM not found: when the VM does not exist in the specified project/zone
//   - Set operation failed: when the GCP API call to attach the schedule policy fails
//
// Example:
//
//	usecase := NewSetSchedulePolicyUseCase(vmRepo)
//	err := usecase.Execute(ctx, "my-project", "us-central1-a", "my-vm", "my-schedule-policy")
//	if err != nil {
//	    log.Fatalf("Failed to set schedule policy: %v", err)
//	}
func (uc *SetSchedulePolicyUseCase) Execute(ctx context.Context, project, zone, name, policyName string) error {
	// 1. VMを取得
	vm, err := uc.vmRepo.FindByName(ctx, project, zone, name)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	// 2. スケジュールポリシー設定実行
	if setErr := uc.vmRepo.SetSchedulePolicy(ctx, vm, policyName); setErr != nil {
		return fmt.Errorf("failed to set schedule policy: %w", setErr)
	}

	return nil
}
