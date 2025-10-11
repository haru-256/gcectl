package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
)

func TestSetSchedulePolicyUseCase_Execute(t *testing.T) {
	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name                  string
		project               string
		zone                  string
		vmName                string
		policyName            string
		mockFindByName        func(ctx context.Context, project, zone, name string) (*model.VM, error)
		mockSetSchedulePolicy func(ctx context.Context, vm *model.VM, policyName string) error
		wantErr               bool
		errContains           string
	}{
		{
			name:       "success: set schedule policy",
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
					SchedulePolicy: "",
				}, nil
			},
			mockSetSchedulePolicy: func(ctx context.Context, vm *model.VM, policyName string) error {
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
			mockSetSchedulePolicy: nil,
			wantErr:               true,
			errContains:           "failed to find VM",
		},
		{
			name:       "error: set operation failed",
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
					SchedulePolicy: "",
				}, nil
			},
			mockSetSchedulePolicy: func(ctx context.Context, vm *model.VM, policyName string) error {
				return errors.New("GCP API error")
			},
			wantErr:     true,
			errContains: "failed to set schedule policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockVMRepository{
				findByNameFunc:        tt.mockFindByName,
				setSchedulePolicyFunc: tt.mockSetSchedulePolicy,
			}

			usecase := NewSetSchedulePolicyUseCase(mockRepo)
			err := usecase.Execute(context.Background(), tt.project, tt.zone, tt.vmName, tt.policyName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Execute() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
