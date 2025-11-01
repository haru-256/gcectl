package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	mock_repository "github.com/haru-256/gcectl/internal/mock/repository"
	"github.com/haru-256/gcectl/internal/usecase/testhelpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUnsetSchedulePolicyUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		project     string
		zone        string
		vmName      string
		policyName  string
		errContains string
		setupMock   func(*mock_repository.MockVMRepository)
		wantErr     bool
	}{
		{
			name:       "success: unset schedule policy",
			project:    "test-project",
			zone:       "us-central1-a",
			vmName:     "test-vm",
			policyName: "my-schedule-policy",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:           "test-vm",
					Project:        "test-project",
					Zone:           "us-central1-a",
					Status:         model.StatusRunning,
					SchedulePolicy: "my-schedule-policy",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, vm, vm, nil))
				m.EXPECT().
					UnsetSchedulePolicy(gomock.Any(), vm, "my-schedule-policy").
					DoAndReturn(func(ctx context.Context, inputVM *model.VM, policyName string) error {
						assert.Equal(t, vm, inputVM)
						assert.Equal(t, "my-schedule-policy", policyName)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:       "error: VM not found",
			project:    "test-project",
			zone:       "us-central1-a",
			vmName:     "nonexistent-vm",
			policyName: "my-schedule-policy",
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
			name:       "error: unset operation failed",
			project:    "test-project",
			zone:       "us-central1-a",
			vmName:     "test-vm",
			policyName: "my-schedule-policy",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:           "test-vm",
					Project:        "test-project",
					Zone:           "us-central1-a",
					Status:         model.StatusRunning,
					SchedulePolicy: "my-schedule-policy",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, vm, vm, nil))
				m.EXPECT().
					UnsetSchedulePolicy(gomock.Any(), vm, "my-schedule-policy").
					DoAndReturn(func(ctx context.Context, inputVM *model.VM, policyName string) error {
						assert.Equal(t, vm, inputVM)
						assert.Equal(t, "my-schedule-policy", policyName)
						return errors.New("GCP API error")
					})
			},
			wantErr:     true,
			errContains: "failed to unset schedule policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockVMRepository(ctrl)
			tt.setupMock(mockRepo)

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
