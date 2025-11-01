package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	mock_repository "github.com/haru-256/gcectl/internal/mock/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var errTestList = errors.New("test error")

//nolint:gocognit // Test function is complex but readable with table-driven design
func TestListVMsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		wantUptimes []string // Expected uptime strings for each VM
		setupMock   func(*mock_repository.MockVMRepository)
		wantLen     int
		wantError   bool
	}{
		{
			name: "single running VM with uptime",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vms := []*model.VM{
					{
						Name:          "test-vm",
						Project:       "test-project",
						Zone:          "us-central1-a",
						MachineType:   "e2-medium",
						Status:        model.StatusRunning,
						LastStartTime: timePtr(time.Now().Add(-2 * time.Hour)),
					},
				}
				m.EXPECT().FindAll(gomock.Any()).Return(vms, nil)
			},
			wantLen:     1,
			wantUptimes: []string{"2h0m0s"}, // Approximately 2 hours
			wantError:   false,
		},
		{
			name: "stopped VM returns N/A uptime",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vms := []*model.VM{
					{
						Name:          "stopped-vm",
						Project:       "test-project",
						Zone:          "us-central1-a",
						MachineType:   "e2-medium",
						Status:        model.StatusStopped,
						LastStartTime: nil,
					},
				}
				m.EXPECT().FindAll(gomock.Any()).Return(vms, nil)
			},
			wantLen:     1,
			wantUptimes: []string{"N/A"},
			wantError:   false,
		},
		{
			name: "multiple VMs with mixed statuses",
			setupMock: func(m *mock_repository.MockVMRepository) {
				vms := []*model.VM{
					{
						Name:          "running-vm",
						Project:       "test-project",
						Zone:          "us-central1-a",
						MachineType:   "e2-medium",
						Status:        model.StatusRunning,
						LastStartTime: timePtr(time.Now().Add(-30 * time.Minute)),
					},
					{
						Name:          "stopped-vm",
						Project:       "test-project",
						Zone:          "us-west1-a",
						MachineType:   "n1-standard-1",
						Status:        model.StatusStopped,
						LastStartTime: nil,
					},
				}
				m.EXPECT().FindAll(gomock.Any()).Return(vms, nil)
			},
			wantLen:     2,
			wantUptimes: []string{"30m0s", "N/A"},
			wantError:   false,
		},
		{
			name: "repository error",
			setupMock: func(m *mock_repository.MockVMRepository) {
				m.EXPECT().FindAll(gomock.Any()).Return(nil, errTestList)
			},
			wantLen:     0,
			wantUptimes: nil,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockVMRepository(ctrl)
			tt.setupMock(mockRepo)

			useCase := NewListVMsUseCase(mockRepo)
			ctx := context.Background()

			items, err := useCase.Execute(ctx)

			// Check error
			if tt.wantError {
				assert.Error(t, err, "Execute() should return an error")
				return
			}

			assert.NoError(t, err, "Execute() should not return an error")

			// Check length
			require.Len(t, items, tt.wantLen, "Execute() should return %d items", tt.wantLen)

			// Check uptime strings
			for i, item := range items {
				// For uptime, check if it's "N/A" or not
				// Detailed format testing is covered in TestFormatUptime
				if tt.wantUptimes[i] == "N/A" {
					assert.Equal(t, "N/A", item.Uptime, "Execute() item[%d].Uptime should be N/A", i)
				} else {
					// For running VMs, just verify it's not "N/A"
					assert.NotEqual(t, "N/A", item.Uptime, "Execute() item[%d].Uptime should not be N/A", i)
				}
			}
		})
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}
