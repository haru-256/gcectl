package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestUnsetSchedulePolicyUseCase_Execute(t *testing.T) {
	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name                    string
		project                 string
		zone                    string
		vmName                  string
		policyName              string
		mockFindByName          func(ctx context.Context, project, zone, name string) (*model.VM, error)
		mockUnsetSchedulePolicy func(ctx context.Context, vm *model.VM, policyName string) error
		wantErr                 bool
		errContains             string
	}{
		{
			name:       "success: unset schedule policy",
			project:    "test-project",
			zone:       "us-central1-a",
			vmName:     "test-vm",
			policyName: "my-schedule-policy",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:           name,
					Project:        project,
					Zone:           zone,
					Status:         model.StatusRunning,
					SchedulePolicy: "my-schedule-policy",
				}, nil
			},
			mockUnsetSchedulePolicy: func(ctx context.Context, vm *model.VM, policyName string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:       "error: VM not found",
			project:    "test-project",
			zone:       "us-central1-a",
			vmName:     "nonexistent-vm",
			policyName: "my-schedule-policy",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return nil, errors.New("VM not found")
			},
			mockUnsetSchedulePolicy: nil,
			wantErr:                 true,
			errContains:             "failed to find VM",
		},
		{
			name:       "error: unset operation failed",
			project:    "test-project",
			zone:       "us-central1-a",
			vmName:     "test-vm",
			policyName: "my-schedule-policy",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:           name,
					Project:        project,
					Zone:           zone,
					Status:         model.StatusRunning,
					SchedulePolicy: "my-schedule-policy",
				}, nil
			},
			mockUnsetSchedulePolicy: func(ctx context.Context, vm *model.VM, policyName string) error {
				return errors.New("GCP API error")
			},
			wantErr:     true,
			errContains: "failed to unset schedule policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockVMRepository{
				findByNameFunc:          tt.mockFindByName,
				unsetSchedulePolicyFunc: tt.mockUnsetSchedulePolicy,
			}

			usecase := NewUnsetSchedulePolicyUseCase(mockRepo)
			err := usecase.Execute(context.Background(), tt.project, tt.zone, tt.vmName, tt.policyName)

			if tt.wantErr {
				assert.Error(t, err, "Execute() should return an error")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "Error should contain %v", tt.errContains)
				}
			} else {
				assert.NoError(t, err, "Execute() should not return an error")
			}
		})
	}
}
