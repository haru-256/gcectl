package usecase

import (
	"context"
	"errors"
	"sync/atomic"
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
		name                string
		configured          []*model.VM
		wantUptimeAvailable []bool
		setupMock           func(*mock_repository.MockVMRepository)
		wantLen             int
		wantError           bool
	}{
		{
			name: "single running VM with uptime",
			configured: []*model.VM{
				{Name: "test-vm", Project: "test-project", Zone: "us-central1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:          "test-vm",
					Project:       "test-project",
					Zone:          "us-central1-a",
					MachineType:   "e2-medium",
					Status:        model.StatusRunning,
					LastStartTime: timePtr(time.Now().Add(-2 * time.Hour)),
				}
				m.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(vm, nil)
			},
			wantLen:             1,
			wantUptimeAvailable: []bool{true},
			wantError:           false,
		},
		{
			name: "stopped VM returns N/A uptime",
			configured: []*model.VM{
				{Name: "stopped-vm", Project: "test-project", Zone: "us-central1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				vm := &model.VM{
					Name:          "stopped-vm",
					Project:       "test-project",
					Zone:          "us-central1-a",
					MachineType:   "e2-medium",
					Status:        model.StatusStopped,
					LastStartTime: nil,
				}
				m.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(vm, nil)
			},
			wantLen:             1,
			wantUptimeAvailable: []bool{false},
			wantError:           false,
		},
		{
			name: "multiple VMs with mixed statuses",
			configured: []*model.VM{
				{Name: "running-vm", Project: "test-project", Zone: "us-central1-a"},
				{Name: "stopped-vm", Project: "test-project", Zone: "us-west1-a"},
			},
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
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					Times(2).
					DoAndReturn(func(ctx context.Context, vm *model.VM) (*model.VM, error) {
						switch vm.Name {
						case "running-vm":
							return vms[0], nil
						case "stopped-vm":
							return vms[1], nil
						default:
							return nil, errors.New("unexpected VM")
						}
					})
			},
			wantLen:             2,
			wantUptimeAvailable: []bool{true, false},
			wantError:           false,
		},
		{
			name: "partial results with repository error",
			configured: []*model.VM{
				{Name: "running-vm", Project: "test-project", Zone: "us-central1-a"},
				{Name: "missing-vm", Project: "test-project", Zone: "us-west1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				runningVM := &model.VM{
					Name:          "running-vm",
					Project:       "test-project",
					Zone:          "us-central1-a",
					MachineType:   "e2-medium",
					Status:        model.StatusRunning,
					LastStartTime: timePtr(time.Now().Add(-30 * time.Minute)),
				}
				m.EXPECT().
					FindByName(gomock.Any(), gomock.Any()).
					Times(2).
					DoAndReturn(func(ctx context.Context, vm *model.VM) (*model.VM, error) {
						switch vm.Name {
						case "running-vm":
							return runningVM, nil
						case "missing-vm":
							return nil, errTestList
						default:
							return nil, errors.New("unexpected VM")
						}
					})
			},
			wantLen:             1,
			wantUptimeAvailable: []bool{true},
			wantError:           true,
		},
		{
			name: "repository error",
			configured: []*model.VM{
				{Name: "error-vm", Project: "test-project", Zone: "us-central1-a"},
			},
			setupMock: func(m *mock_repository.MockVMRepository) {
				m.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(nil, errTestList)
			},
			wantLen:             0,
			wantUptimeAvailable: nil,
			wantError:           true,
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

			items, err := useCase.Execute(ctx, tt.configured)

			if tt.wantError {
				assert.Error(t, err, "Execute() should return an error")
			} else {
				assert.NoError(t, err, "Execute() should not return an error")
			}

			require.Len(t, items, tt.wantLen, "Execute() should return %d items", tt.wantLen)

			for i, item := range items {
				if tt.wantUptimeAvailable[i] {
					assert.NotEqual(t, "N/A", item.Uptime, "Execute() item[%d].Uptime should contain uptime", i)
				} else {
					assert.Equal(t, "N/A", item.Uptime, "Execute() item[%d].Uptime should be N/A", i)
				}
			}
		})
	}
}

func TestListVMsUseCase_ExecuteLimitsConcurrentLookups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configured := make([]*model.VM, maxConcurrentVMLookups+1)
	for i := range configured {
		configured[i] = &model.VM{Name: "test-vm", Project: "test-project", Zone: "us-central1-a"}
	}

	var inFlight int32
	var maxInFlight int32
	release := make(chan struct{})
	mockRepo := mock_repository.NewMockVMRepository(ctrl)
	mockRepo.EXPECT().
		FindByName(gomock.Any(), gomock.Any()).
		Times(len(configured)).
		DoAndReturn(func(ctx context.Context, vm *model.VM) (*model.VM, error) {
			current := atomic.AddInt32(&inFlight, 1)
			for {
				previous := atomic.LoadInt32(&maxInFlight)
				if current <= previous || atomic.CompareAndSwapInt32(&maxInFlight, previous, current) {
					break
				}
			}
			<-release
			atomic.AddInt32(&inFlight, -1)
			return &model.VM{
				Name:          vm.Name,
				Project:       vm.Project,
				Zone:          vm.Zone,
				MachineType:   "e2-medium",
				Status:        model.StatusRunning,
				LastStartTime: timePtr(time.Now().Add(-30 * time.Minute)),
			}, nil
		})

	done := make(chan error, 1)
	go func() {
		_, err := NewListVMsUseCase(mockRepo).Execute(context.Background(), configured)
		done <- err
	}()

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&maxInFlight) == maxConcurrentVMLookups
	}, time.Second, 10*time.Millisecond)

	for range configured {
		release <- struct{}{}
	}

	require.NoError(t, <-done)
	assert.LessOrEqual(t, atomic.LoadInt32(&maxInFlight), int32(maxConcurrentVMLookups))
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}
