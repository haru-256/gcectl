package usecase

import (
	"testing"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
)

func TestCalculateUptimeString(t *testing.T) {
	//nolint:govet // Test struct prioritizes readability over field alignment
	tests := []struct {
		name   string
		vm     *model.VM
		now    time.Time
		want   string
		wantNA bool // True if we expect "N/A"
	}{
		{
			name: "running VM with valid uptime",
			vm: &model.VM{
				Name:   "test-vm",
				Status: model.StatusRunning,
				LastStartTime: func() *time.Time {
					// Started 2 hours ago
					t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
			now:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			want:   "2h0m0s",
			wantNA: false,
		},
		{
			name: "stopped VM returns N/A",
			vm: &model.VM{
				Name:          "stopped-vm",
				Status:        model.StatusStopped,
				LastStartTime: nil,
			},
			now:    time.Now(),
			want:   "N/A",
			wantNA: true,
		},
		{
			name: "running VM without LastStartTime returns N/A",
			vm: &model.VM{
				Name:          "running-vm-no-start",
				Status:        model.StatusRunning,
				LastStartTime: nil,
			},
			now:    time.Now(),
			want:   "N/A",
			wantNA: true,
		},
		{
			name: "terminated VM returns N/A",
			vm: &model.VM{
				Name:   "terminated-vm",
				Status: model.StatusTerminated,
			},
			now:    time.Now(),
			want:   "N/A",
			wantNA: true,
		},
		{
			name: "provisioning VM returns N/A",
			vm: &model.VM{
				Name:   "provisioning-vm",
				Status: model.StatusProvisioning,
			},
			now:    time.Now(),
			want:   "N/A",
			wantNA: true,
		},
		{
			name: "running VM with 30 minutes uptime",
			vm: &model.VM{
				Name:   "test-vm-30m",
				Status: model.StatusRunning,
				LastStartTime: func() *time.Time {
					t := time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)
					return &t
				}(),
			},
			now:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			want:   "30m0s",
			wantNA: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateUptimeString(tt.vm, tt.now)

			if tt.wantNA {
				if got != "N/A" {
					t.Errorf("calculateUptimeString() = %v, want N/A", got)
				}
			} else {
				if got != tt.want {
					t.Errorf("calculateUptimeString() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
