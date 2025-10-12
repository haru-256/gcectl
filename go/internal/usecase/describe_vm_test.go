package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

var errTestDescribe = errors.New("test error")

//nolint:govet // Test structs prioritize readability over field alignment
type mockVMRepositoryForDescribe struct {
	findByNameFunc func(ctx context.Context, project, zone, name string) (*model.VM, error)
}

func (m *mockVMRepositoryForDescribe) FindByName(ctx context.Context, project, zone, name string) (*model.VM, error) {
	return m.findByNameFunc(ctx, project, zone, name)
}

func (m *mockVMRepositoryForDescribe) List(ctx context.Context, project, zone string) ([]*model.VM, error) {
	return nil, errors.New("not implemented")
}

func (m *mockVMRepositoryForDescribe) FindAll(ctx context.Context) ([]*model.VM, error) {
	return nil, errors.New("not implemented")
}

func (m *mockVMRepositoryForDescribe) Start(ctx context.Context, vm *model.VM) error {
	return errors.New("not implemented")
}

func (m *mockVMRepositoryForDescribe) Stop(ctx context.Context, vm *model.VM) error {
	return errors.New("not implemented")
}

func (m *mockVMRepositoryForDescribe) SetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
	return errors.New("not implemented")
}

func (m *mockVMRepositoryForDescribe) UnsetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
	return errors.New("not implemented")
}

func (m *mockVMRepositoryForDescribe) UpdateMachineType(ctx context.Context, vm *model.VM, machineType string) error {
	return errors.New("not implemented")
}

//nolint:gocognit // Test function is complex but readable with table-driven design
func TestDescribeVM(t *testing.T) {
	//nolint:govet // Test struct prioritizes readability over field alignment
	tests := []struct {
		name           string
		project        string
		zone           string
		vmName         string
		mockFindByName func(ctx context.Context, project, zone, name string) (*model.VM, error)
		wantVM         *model.VM
		wantUptime     string
		wantErr        bool
	}{
		{
			name:    "running VM with uptime",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "test-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:        "test-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					MachineType: "e2-medium",
					Status:      model.StatusRunning,
					LastStartTime: func() *time.Time {
						t := time.Now().Add(-2 * time.Hour)
						return &t
					}(),
				}, nil
			},
			wantVM: &model.VM{
				Name:        "test-vm",
				Project:     "test-project",
				Zone:        "us-central1-a",
				MachineType: "e2-medium",
				Status:      model.StatusRunning,
			},
			wantUptime: "", // We'll check this is not "N/A" in the test
			wantErr:    false,
		},
		{
			name:    "stopped VM",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "stopped-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return &model.VM{
					Name:        "stopped-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					MachineType: "e2-medium",
					Status:      model.StatusStopped,
				}, nil
			},
			wantVM: &model.VM{
				Name:        "stopped-vm",
				Project:     "test-project",
				Zone:        "us-central1-a",
				MachineType: "e2-medium",
				Status:      model.StatusStopped,
			},
			wantUptime: "N/A",
			wantErr:    false,
		},
		{
			name:    "repository error",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "error-vm",
			mockFindByName: func(ctx context.Context, project, zone, name string) (*model.VM, error) {
				return nil, errTestDescribe
			},
			wantVM:     nil,
			wantUptime: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVMRepositoryForDescribe{
				findByNameFunc: tt.mockFindByName,
			}

			vm, uptime, err := DescribeVM(context.Background(), repo, tt.project, tt.zone, tt.vmName)

			if tt.wantErr {
				assert.Error(t, err, "DescribeVM() should return an error")
				return
			}

			assert.NoError(t, err, "DescribeVM() should not return an error")

			// Check VM fields (except LastStartTimestamp which varies)
			assert.Equal(t, tt.wantVM.Name, vm.Name, "VM.Name should match")
			assert.Equal(t, tt.wantVM.Project, vm.Project, "VM.Project should match")
			assert.Equal(t, tt.wantVM.Zone, vm.Zone, "VM.Zone should match")
			assert.Equal(t, tt.wantVM.MachineType, vm.MachineType, "VM.MachineType should match")
			assert.Equal(t, tt.wantVM.Status, vm.Status, "VM.Status should match")

			// Check uptime
			if tt.wantUptime == "N/A" {
				assert.Equal(t, "N/A", uptime, "Uptime should be N/A")
			} else if tt.name == "running VM with uptime" {
				// For running VM, check that uptime is not "N/A" and contains time components
				assert.NotEqual(t, "N/A", uptime, "Uptime should not be N/A")
			}
		})
	}
}
