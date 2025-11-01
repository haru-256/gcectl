package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
	mock_repository "github.com/haru-256/gcectl/internal/mock/repository"
	"github.com/haru-256/gcectl/internal/usecase/testhelpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var loggerForUpdateMachineType = log.NewLogger()

func TestUpdateMachineTypeUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		project     string
		zone        string
		vmName      string
		machineType string
		errContains string
		setupMock   func(*mock_repository.MockVMRepository)
		wantErr     bool
	}{
		{
			name:        "success: update machine type of stopped VM",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "test-vm",
			machineType: "e2-medium",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:        "test-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					Status:      model.StatusStopped,
					MachineType: "e2-small",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, vm, vm, nil))
				m.EXPECT().
					UpdateMachineType(gomock.Any(), vm, "e2-medium").
					DoAndReturn(func(ctx context.Context, inputVM *model.VM, machineType string) error {
						assert.Equal(t, vm, inputVM)
						assert.Equal(t, "e2-medium", machineType)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:        "success: update machine type of terminated VM",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "test-vm",
			machineType: "n1-standard-1",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:        "test-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					Status:      model.StatusTerminated,
					MachineType: "e2-small",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, vm, vm, nil))
				m.EXPECT().
					UpdateMachineType(gomock.Any(), vm, "n1-standard-1").
					DoAndReturn(func(ctx context.Context, inputVM *model.VM, machineType string) error {
						assert.Equal(t, vm, inputVM)
						assert.Equal(t, "n1-standard-1", machineType)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:        "error: VM not found",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "nonexistent-vm",
			machineType: "e2-medium",
			setupMock: func(m *mock_repository.MockVMRepository) {
				expectedVM := &model.VM{
					Name:    "nonexistent-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, expectedVM, nil, errors.New("VM not found")))
			},
			wantErr:     true,
			errContains: "failed to find VM",
		},
		{
			name:        "error: VM is running",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "running-vm",
			machineType: "e2-medium",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:        "running-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					Status:      model.StatusRunning,
					MachineType: "e2-small",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, vm, vm, nil))
			},
			wantErr:     true,
			errContains: "must be stopped",
		},
		{
			name:        "error: update operation failed",
			project:     "test-project",
			zone:        "us-central1-a",
			vmName:      "test-vm",
			machineType: "e2-medium",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:        "test-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					Status:      model.StatusStopped,
					MachineType: "e2-small",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, vm, vm, nil))
				m.EXPECT().
					UpdateMachineType(gomock.Any(), vm, "e2-medium").
					DoAndReturn(func(ctx context.Context, inputVM *model.VM, machineType string) error {
						assert.Equal(t, vm, inputVM)
						assert.Equal(t, "e2-medium", machineType)
						return errors.New("GCP API error")
					})
			},
			wantErr:     true,
			errContains: "failed to update machine type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockVMRepository(ctrl)
			tt.setupMock(mockRepo)

			usecase := NewUpdateMachineTypeUseCase(mockRepo, loggerForUpdateMachineType)
			err := usecase.Execute(context.Background(), tt.project, tt.zone, tt.vmName, tt.machineType)

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
