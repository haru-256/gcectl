package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestStopVMUseCase_Execute(t *testing.T) {
	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name           string
		project        string
		zone           string
		vmName         string
		mockFindByName func(ctx context.Context, project, zone, name string) (*model.VM, error)
		mockStop       func(ctx context.Context, vm *model.VM) error
		wantErr        bool
		errContains    string
	}{
		{
			name:    "success: stop running VM",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "test-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusRunning,
				}, nil
			},
			mockStop: func(ctx context.Context, vm *model.VM) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "error: VM not found",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "nonexistent-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return nil, errors.New("VM not found")
			},
			mockStop:    nil,
			wantErr:     true,
			errContains: "failed to find VM",
		},
		{
			name:    "error: VM is already stopped",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "stopped-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusStopped,
				}, nil
			},
			mockStop:    nil,
			wantErr:     true,
			errContains: "cannot be stopped",
		},
		{
			name:    "error: VM is terminated",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "terminated-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusTerminated,
				}, nil
			},
			mockStop:    nil,
			wantErr:     true,
			errContains: "cannot be stopped",
		},
		{
			name:    "error: stop operation failed",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "test-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusRunning,
				}, nil
			},
			mockStop: func(ctx context.Context, vm *model.VM) error {
				return errors.New("GCP API error")
			},
			wantErr:     true,
			errContains: "failed to stop VM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockVMRepository{
				findByNameFunc: tt.mockFindByName,
				stopFunc:       tt.mockStop,
			}

			usecase := NewStopVMUseCase(mockRepo)
			err := usecase.Execute(context.Background(), tt.project, tt.zone, tt.vmName)

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
