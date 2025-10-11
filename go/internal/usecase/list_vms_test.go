package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
)

var errTestList = errors.New("test error")

//nolint:gocognit // Test function is complex but readable with table-driven design
func TestListVMsUseCase_Execute(t *testing.T) {
	//nolint:govet // Field alignment is less important than readability in test structs
	tests := []struct {
		name        string
		mockVMs     []*model.VM
		mockError   error
		wantLen     int
		wantUptimes []string // Expected uptime strings for each VM
		wantError   bool
	}{
		{
			name: "single running VM with uptime",
			mockVMs: []*model.VM{
				{
					Name:          "test-vm",
					Project:       "test-project",
					Zone:          "us-central1-a",
					MachineType:   "e2-medium",
					Status:        model.StatusRunning,
					LastStartTime: timePtr(time.Now().Add(-2 * time.Hour)),
				},
			},
			mockError:   nil,
			wantLen:     1,
			wantUptimes: []string{"2h0m0s"}, // Approximately 2 hours
			wantError:   false,
		},
		{
			name: "stopped VM returns N/A uptime",
			mockVMs: []*model.VM{
				{
					Name:          "stopped-vm",
					Project:       "test-project",
					Zone:          "us-central1-a",
					MachineType:   "e2-medium",
					Status:        model.StatusStopped,
					LastStartTime: nil,
				},
			},
			mockError:   nil,
			wantLen:     1,
			wantUptimes: []string{"N/A"},
			wantError:   false,
		},
		{
			name: "multiple VMs with mixed statuses",
			mockVMs: []*model.VM{
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
			},
			mockError:   nil,
			wantLen:     2,
			wantUptimes: []string{"30m0s", "N/A"},
			wantError:   false,
		},
		{
			name:        "repository error",
			mockVMs:     nil,
			mockError:   errTestList,
			wantLen:     0,
			wantUptimes: nil,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVMRepository{
				findAllFunc: func(ctx context.Context) ([]*model.VM, error) {
					return tt.mockVMs, tt.mockError
				},
			}

			useCase := NewListVMsUseCase(repo)
			ctx := context.Background()

			items, err := useCase.Execute(ctx)

			// Check error
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			// Check length
			if len(items) != tt.wantLen {
				t.Errorf("Execute() returned %d items, want %d", len(items), tt.wantLen)
				return
			}

			// Check VM data and uptime strings
			for i, item := range items {
				if item.VM != tt.mockVMs[i] {
					t.Errorf("Execute() item[%d].VM = %v, want %v", i, item.VM, tt.mockVMs[i])
				}

				// For uptime, we need to be flexible with time-based tests
				// Check if it's "N/A" or a valid duration string
				if tt.wantUptimes[i] == "N/A" {
					if item.Uptime != "N/A" {
						t.Errorf("Execute() item[%d].Uptime = %s, want N/A", i, item.Uptime)
					}
				} else {
					// For running VMs, just verify it's not "N/A" and is a valid duration format
					if item.Uptime == "N/A" {
						t.Errorf("Execute() item[%d].Uptime = N/A, want a valid duration", i)
					}
					// Verify it's a parseable duration
					if _, parseErr := time.ParseDuration(item.Uptime); parseErr != nil {
						t.Errorf("Execute() item[%d].Uptime = %s is not a valid duration: %v", i, item.Uptime, parseErr)
					}
				}
			}
		})
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}
