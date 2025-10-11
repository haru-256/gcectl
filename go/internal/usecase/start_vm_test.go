package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
)

// mockVMRepository is a mock implementation of repository.VMRepository for testing
type mockVMRepository struct {
	findByNameFunc          func(ctx context.Context, project, zone, name string) (*model.VM, error)
	findAllFunc             func(ctx context.Context) ([]*model.VM, error)
	startFunc               func(ctx context.Context, vm *model.VM) error
	stopFunc                func(ctx context.Context, vm *model.VM) error
	updateMachineTypeFunc   func(ctx context.Context, vm *model.VM, machineType string) error
	setSchedulePolicyFunc   func(ctx context.Context, vm *model.VM, policyName string) error
	unsetSchedulePolicyFunc func(ctx context.Context, vm *model.VM, policyName string) error
}

func (m *mockVMRepository) FindByName(ctx context.Context, project, zone, name string) (*model.VM, error) {
	if m.findByNameFunc != nil {
		return m.findByNameFunc(ctx, project, zone, name)
	}
	return nil, errors.New("not implemented")
}

func (m *mockVMRepository) FindAll(ctx context.Context) ([]*model.VM, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockVMRepository) Start(ctx context.Context, vm *model.VM) error {
	if m.startFunc != nil {
		return m.startFunc(ctx, vm)
	}
	return errors.New("not implemented")
}

func (m *mockVMRepository) Stop(ctx context.Context, vm *model.VM) error {
	if m.stopFunc != nil {
		return m.stopFunc(ctx, vm)
	}
	return errors.New("not implemented")
}

func (m *mockVMRepository) UpdateMachineType(ctx context.Context, vm *model.VM, machineType string) error {
	if m.updateMachineTypeFunc != nil {
		return m.updateMachineTypeFunc(ctx, vm, machineType)
	}
	return errors.New("not implemented")
}

func (m *mockVMRepository) SetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
	if m.setSchedulePolicyFunc != nil {
		return m.setSchedulePolicyFunc(ctx, vm, policyName)
	}
	return errors.New("not implemented")
}

func (m *mockVMRepository) UnsetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
	if m.unsetSchedulePolicyFunc != nil {
		return m.unsetSchedulePolicyFunc(ctx, vm, policyName)
	}
	return errors.New("not implemented")
}

func TestStartVMUseCase_Execute(t *testing.T) {
	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name           string
		project        string
		zone           string
		vmName         string
		mockFindByName func(ctx context.Context, project, zone, name string) (*model.VM, error)
		mockStart      func(ctx context.Context, vm *model.VM) error
		wantErr        bool
		errContains    string
	}{
		{
			name:    "success: start stopped VM",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "test-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusStopped,
				}, nil
			},
			mockStart: func(ctx context.Context, vm *model.VM) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "success: start terminated VM",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "test-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusTerminated,
				}, nil
			},
			mockStart: func(ctx context.Context, vm *model.VM) error {
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
			mockStart:   nil,
			wantErr:     true,
			errContains: "failed to find VM",
		},
		{
			name:    "error: VM is already running",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "running-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusRunning,
				}, nil
			},
			mockStart:   nil,
			wantErr:     true,
			errContains: "cannot be started",
		},
		{
			name:    "error: start operation failed",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "test-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:    name,
					Project: project,
					Zone:    zone,
					Status:  model.StatusStopped,
				}, nil
			},
			mockStart: func(ctx context.Context, vm *model.VM) error {
				return errors.New("GCP API error")
			},
			wantErr:     true,
			errContains: "failed to start VM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockVMRepository{
				findByNameFunc: tt.mockFindByName,
				startFunc:      tt.mockStart,
			}

			usecase := NewStartVMUseCase(mockRepo)
			err := usecase.Execute(context.Background(), tt.project, tt.zone, tt.vmName)

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

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
