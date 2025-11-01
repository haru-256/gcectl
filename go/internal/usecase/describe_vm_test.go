package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	mock_repository "github.com/haru-256/gcectl/internal/mock/repository"
	"github.com/haru-256/gcectl/internal/usecase/testhelpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var errTestDescribe = errors.New("test error")

//nolint:gocognit // Test function is complex but readable with table-driven design
func TestDescribeVM(t *testing.T) {
	tests := []struct {
		name       string
		project    string
		zone       string
		vmName     string
		setupMock  func(*mock_repository.MockVMRepository)
		wantVM     *model.VM
		wantUptime string
		wantErr    bool
	}{
		{
			name:    "running VM with uptime",
			project: "test-project",
			zone:    "us-central1-a",
			vmName:  "test-vm",
			setupMock: func(m *mock_repository.MockVMRepository) {
				expectedVM := &model.VM{
					Name:    "test-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
				}
				returnVM := &model.VM{
					Name:        "test-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					MachineType: "e2-medium",
					Status:      model.StatusRunning,
					LastStartTime: func() *time.Time {
						t := time.Now().Add(-2 * time.Hour)
						return &t
					}(),
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, expectedVM, returnVM, nil))
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
			setupMock: func(m *mock_repository.MockVMRepository) {
				expectedVM := &model.VM{
					Name:    "stopped-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
				}
				returnVM := &model.VM{
					Name:        "stopped-vm",
					Project:     "test-project",
					Zone:        "us-central1-a",
					MachineType: "e2-medium",
					Status:      model.StatusStopped,
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, expectedVM, returnVM, nil))
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
			setupMock: func(m *mock_repository.MockVMRepository) {
				expectedVM := &model.VM{
					Name:    "error-vm",
					Project: "test-project",
					Zone:    "us-central1-a",
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					DoAndReturn(testhelpers.VMFindByNameMatcher(t, expectedVM, nil, errTestDescribe))
			},
			wantVM:     nil,
			wantUptime: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockVMRepository(ctrl)
			tt.setupMock(mockRepo)

			vm, uptime, err := DescribeVM(context.Background(), mockRepo, tt.project, tt.zone, tt.vmName)

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
