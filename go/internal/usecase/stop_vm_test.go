package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	mock_repository "github.com/haru-256/gcectl/internal/mock/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStopVMUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		vms         []*model.VM
		errContains string
		setupMock   func(*mock_repository.MockVMRepository)
		wantErr     bool
	}{
		{
			name: "success: stop single running VM",
			vms: []*model.VM{
				{Name: "test-vm", Project: "test-project", Zone: "us-central1-a", Status: model.StatusRunning},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:    "test-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
					Status:  model.StatusRunning,
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						assert.Equal(t, "test-project", inputVM.Project)
						assert.Equal(t, "us-central1-a", inputVM.Zone)
						assert.Equal(t, "test-vm", inputVM.Name)
						return vm, nil
					})
				m.EXPECT().
					Stop(gomock.Any(), vm).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						assert.Equal(t, vm, inputVM)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "success: stop multiple running VMs in parallel",
			vms: []*model.VM{
				{Name: "vm1", Project: "project1", Zone: "zone1", Status: model.StatusRunning},
				{Name: "vm2", Project: "project2", Zone: "zone2", Status: model.StatusRunning},
				{Name: "vm3", Project: "project3", Zone: "zone3", Status: model.StatusRunning},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm1 := &model.VM{Name: "vm1", Project: "project1", Zone: "zone1", Status: model.StatusRunning}
				vm2 := &model.VM{Name: "vm2", Project: "project2", Zone: "zone2", Status: model.StatusRunning}
				vm3 := &model.VM{Name: "vm3", Project: "project3", Zone: "zone3", Status: model.StatusRunning}

				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						switch inputVM.Name {
						case "vm1":
							assert.Equal(t, "project1", inputVM.Project)
							assert.Equal(t, "zone1", inputVM.Zone)
							return vm1, nil
						case "vm2":
							assert.Equal(t, "project2", inputVM.Project)
							assert.Equal(t, "zone2", inputVM.Zone)
							return vm2, nil
						case "vm3":
							assert.Equal(t, "project3", inputVM.Project)
							assert.Equal(t, "zone3", inputVM.Zone)
							return vm3, nil
						default:
							t.Errorf("unexpected VM name: %s", inputVM.Name)
							return nil, errors.New("unexpected VM")
						}
					}).
					Times(3)
				m.EXPECT().
					Stop(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						if inputVM.Name != "vm1" && inputVM.Name != "vm2" && inputVM.Name != "vm3" {
							t.Errorf("unexpected VM in Stop: %s", inputVM.Name)
						}
						return nil
					}).
					Times(3)
			},
			wantErr: false,
		},
		{
			name: "error: VM not found",
			vms: []*model.VM{
				{Name: "nonexistent-vm", Project: "test-project", Zone: "us-central1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						assert.Equal(t, "test-project", inputVM.Project)
						assert.Equal(t, "us-central1-a", inputVM.Zone)
						assert.Equal(t, "nonexistent-vm", inputVM.Name)
						return nil, errors.New("VM not found")
					})
			},
			wantErr:     true,
			errContains: "failed to find",
		},
		{
			name: "error: VM returns nil without error",
			vms: []*model.VM{
				{Name: "test-vm", Project: "test-project", Zone: "us-central1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						assert.Equal(t, "test-project", inputVM.Project)
						assert.Equal(t, "us-central1-a", inputVM.Zone)
						assert.Equal(t, "test-vm", inputVM.Name)
						return nil, nil
					})
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "error: VM is already stopped",
			vms: []*model.VM{
				{Name: "stopped-vm", Project: "test-project", Zone: "us-central1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:    "stopped-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
					Status:  model.StatusStopped,
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						assert.Equal(t, "test-project", inputVM.Project)
						assert.Equal(t, "us-central1-a", inputVM.Zone)
						assert.Equal(t, "stopped-vm", inputVM.Name)
						return vm, nil
					})
			},
			wantErr:     true,
			errContains: "cannot be stopped",
		},
		{
			name: "error: stop operation failed",
			vms: []*model.VM{
				{Name: "test-vm", Project: "test-project", Zone: "us-central1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:    "test-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
					Status:  model.StatusRunning,
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						assert.Equal(t, "test-project", inputVM.Project)
						assert.Equal(t, "us-central1-a", inputVM.Zone)
						assert.Equal(t, "test-vm", inputVM.Name)
						return vm, nil
					})
				m.EXPECT().
					Stop(gomock.Any(), vm).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						assert.Equal(t, vm, inputVM)
						return errors.New("GCP API error")
					})
			},
			wantErr:     true,
			errContains: "failed to stop",
		},
		{
			name: "error: fail-fast behavior - one VM fails, all stop",
			vms: []*model.VM{
				{Name: "vm1", Project: "project1", Zone: "zone1"},
				{Name: "vm2", Project: "project2", Zone: "zone2"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						if inputVM.Name == "vm1" {
							assert.Equal(t, "project1", inputVM.Project)
							assert.Equal(t, "zone1", inputVM.Zone)
							return nil, errors.New("VM1 not found")
						}
						if inputVM.Name == "vm2" {
							assert.Equal(t, "project2", inputVM.Project)
							assert.Equal(t, "zone2", inputVM.Zone)
							return &model.VM{Name: "vm2", Project: "project2", Zone: "zone2", Status: model.StatusRunning}, nil
						}
						t.Errorf("unexpected VM name: %s", inputVM.Name)
						return nil, errors.New("unexpected VM")
					}).
					AnyTimes()
				m.EXPECT().
					Stop(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						assert.Equal(t, "vm2", inputVM.Name)
						return nil
					}).
					AnyTimes()
			},
			wantErr:     true,
			errContains: "VM1 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockVMRepository(ctrl)
			tt.setupMock(mockRepo)

			usecase := NewStopVMUseCase(mockRepo)
			err := usecase.Execute(context.Background(), tt.vms)

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
