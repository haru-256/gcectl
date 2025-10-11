package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
)

func TestUpdateMachineTypeUseCase_Execute(t *testing.T) {
	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name                  string
		project               string
		zone                  string
		vmName                string
		machineType           string
		mockFindByName        func(ctx context.Context, project, zone, name string) (*model.VM, error)
		mockUpdateMachineType func(ctx context.Context, vm *model.VM, machineType string) error
		wantErr               bool
		errContains           string
	}{
		{
			name:        "success: update machine type of stopped VM",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "test-vm",
			machineType: "e2-medium",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:        name,
					Project:     project,
					Zone:        zone,
					Status:      model.StatusStopped,
					MachineType: "e2-small",
				}, nil
			},
			mockUpdateMachineType: func(ctx context.Context, vm *model.VM, machineType string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:        "success: update machine type of terminated VM",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "test-vm",
			machineType: "n1-standard-1",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:        name,
					Project:     project,
					Zone:        zone,
					Status:      model.StatusTerminated,
					MachineType: "e2-small",
				}, nil
			},
			mockUpdateMachineType: func(ctx context.Context, vm *model.VM, machineType string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:        "error: VM not found",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "nonexistent-vm",
			machineType: "e2-medium",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return nil, errors.New("VM not found")
			},
			mockUpdateMachineType: nil,
			wantErr:               true,
			errContains:           "failed to find VM",
		},
		{
			name:        "error: VM is running",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "running-vm",
			machineType: "e2-medium",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:        name,
					Project:     project,
					Zone:        zone,
					Status:      model.StatusRunning,
					MachineType: "e2-small",
				}, nil
			},
			mockUpdateMachineType: nil,
			wantErr:               true,
			errContains:           "must be stopped",
		},
		{
			name:        "error: update operation failed",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "test-vm",
			machineType: "e2-medium",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:        name,
					Project:     project,
					Zone:        zone,
					Status:      model.StatusStopped,
					MachineType: "e2-small",
				}, nil
			},
			mockUpdateMachineType: func(ctx context.Context, vm *model.VM, machineType string) error {
				return errors.New("GCP API error")
			},
			wantErr:     true,
			errContains: "failed to update machine type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockVMRepository{
				findByNameFunc:        tt.mockFindByName,
				updateMachineTypeFunc: tt.mockUpdateMachineType,
			}

			usecase := NewUpdateMachineTypeUseCase(mockRepo)
			err := usecase.Execute(context.Background(), tt.project, tt.zone, tt.vmName, tt.machineType)

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
