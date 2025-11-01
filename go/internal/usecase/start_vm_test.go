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

func TestStartVMUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		vms         []*model.VM
		errContains string
		setupMock   func(*mock_repository.MockVMRepository)
		wantErr     bool
	}{
		{
			name: "success: start single stopped VM",
			vms: []*model.VM{
				{Project: "test-project", Zone: "us-central1-a", Name: "test-vm"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:    "test-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
					Status:  model.StatusStopped,
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
					Start(gomock.Any(), vm).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						assert.Equal(t, vm, inputVM)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "success: start multiple VMs in parallel",
			vms: []*model.VM{
				{Project: "test-project", Zone: "us-central1-a", Name: "vm-1"},
				{Project: "test-project", Zone: "us-west1-a", Name: "vm-2"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm1 := &model.VM{
					Name:    "vm-1",
					Project: "test-project",
					Zone:    "us-central1-a",
					Status:  model.StatusStopped,
				}
				vm2 := &model.VM{
					Name:    "vm-2",
					Project: "test-project",
					Zone:    "us-west1-a",
					Status:  model.StatusTerminated,
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						if inputVM.Name == "vm-1" {
							assert.Equal(t, "test-project", inputVM.Project)
							assert.Equal(t, "us-central1-a", inputVM.Zone)
							return vm1, nil
						}
						if inputVM.Name == "vm-2" {
							assert.Equal(t, "test-project", inputVM.Project)
							assert.Equal(t, "us-west1-a", inputVM.Zone)
							return vm2, nil
						}
						t.Errorf("unexpected VM name: %s", inputVM.Name)
						return nil, errors.New("unexpected VM")
					}).
					Times(2)
				m.EXPECT().
					Start(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						if inputVM.Name != "vm-1" && inputVM.Name != "vm-2" {
							t.Errorf("unexpected VM in Start: %s", inputVM.Name)
						}
						return nil
					}).
					Times(2)
			},
			wantErr: false,
		},
		{
			name: "error: VM not found",
			vms: []*model.VM{
				{Project: "test-project", Zone: "us-central1-a", Name: "nonexistent-vm"},
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
				{Project: "test-project", Zone: "us-central1-a", Name: "nil-vm"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						assert.Equal(t, "test-project", inputVM.Project)
						assert.Equal(t, "us-central1-a", inputVM.Zone)
						assert.Equal(t, "nil-vm", inputVM.Name)
						return nil, nil
					})
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "error: VM is already running",
			vms: []*model.VM{
				{Project: "test-project", Zone: "us-central1-a", Name: "running-vm"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:    "running-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
					Status:  model.StatusRunning,
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						assert.Equal(t, "test-project", inputVM.Project)
						assert.Equal(t, "us-central1-a", inputVM.Zone)
						assert.Equal(t, "running-vm", inputVM.Name)
						return vm, nil
					})
			},
			wantErr:     true,
			errContains: "cannot be started",
		},
		{
			name: "error: start operation failed",
			vms: []*model.VM{
				{Project: "test-project", Zone: "us-central1-a", Name: "test-vm"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:    "test-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
					Status:  model.StatusStopped,
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
					Start(gomock.Any(), vm).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						assert.Equal(t, vm, inputVM)
						return errors.New("GCP API error")
					})
			},
			wantErr:     true,
			errContains: "failed to start",
		},
		{
			name: "error: one VM fails, entire operation fails (fail-fast)",
			vms: []*model.VM{
				{Project: "test-project", Zone: "us-central1-a", Name: "vm-1"},
				{Project: "test-project", Zone: "us-west1-a", Name: "vm-2"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
						if inputVM.Name == "vm-1" {
							assert.Equal(t, "test-project", inputVM.Project)
							assert.Equal(t, "us-central1-a", inputVM.Zone)
							return &model.VM{
								Name:    "vm-1",
								Project: "test-project",
								Zone:    "us-central1-a",
								Status:  model.StatusStopped,
							}, nil
						}
						if inputVM.Name == "vm-2" {
							assert.Equal(t, "test-project", inputVM.Project)
							assert.Equal(t, "us-west1-a", inputVM.Zone)
							return nil, errors.New("VM not found")
						}
						t.Errorf("unexpected VM name: %s", inputVM.Name)
						return nil, errors.New("unexpected VM")
					}).
					AnyTimes()
				m.EXPECT().
					Start(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inputVM *model.VM) error {
						assert.Equal(t, "vm-1", inputVM.Name)
						return nil
					}).
					AnyTimes()
			},
			wantErr:     true,
			errContains: "failed to find",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockVMRepository(ctrl)
			tt.setupMock(mockRepo)

			usecase := NewStartVMUseCase(mockRepo)
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
