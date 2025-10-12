package usecase

import (
	"testing"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestCalculateUptimeString(t *testing.T) {
	tests := []struct {
		vm     *model.VM
		now    time.Time
		name   string
		want   string
		wantNA bool
	}{
		{
			name: "running VM with valid uptime (hours)",
			vm: &model.VM{
				Name:   "test-vm",
				Status: model.StatusRunning,
				LastStartTime: func() *time.Time {
					// Started 2 hours and 30 minutes ago
					t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
			now:    time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
			want:   "2h30m",
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
			name: "running VM with minutes and seconds",
			vm: &model.VM{
				Name:   "test-vm-minutes",
				Status: model.StatusRunning,
				LastStartTime: func() *time.Time {
					// Started 5 minutes and 30 seconds ago
					t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
			now:    time.Date(2024, 1, 1, 10, 5, 30, 0, time.UTC),
			want:   "5m30s",
			wantNA: false,
		},
		{
			name: "running VM with seconds only",
			vm: &model.VM{
				Name:   "test-vm-seconds",
				Status: model.StatusRunning,
				LastStartTime: func() *time.Time {
					// Started 45 seconds ago
					t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
			now:    time.Date(2024, 1, 1, 10, 0, 45, 0, time.UTC),
			want:   "45s",
			wantNA: false,
		},
		{
			name: "running VM with 1 minute 5 seconds (rounds to seconds)",
			vm: &model.VM{
				Name:   "test-vm-1m5s",
				Status: model.StatusRunning,
				LastStartTime: func() *time.Time {
					// Started 1 minute 5.6 seconds ago
					t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
					return &t
				}(),
			},
			now:    time.Date(2024, 1, 1, 10, 1, 5, 600000000, time.UTC), // 5.6 seconds
			want:   "1m6s",                                               // Rounded to 6s
			wantNA: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateUptimeString(tt.vm, tt.now)

			if tt.wantNA {
				assert.Equal(t, "N/A", got, "calculateUptimeString() should return N/A")
			} else {
				assert.Equal(t, tt.want, got, "calculateUptimeString() should return %v", tt.want)
			}
		})
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		duration time.Duration
		name     string
		want     string
	}{
		{
			name:     "days, hours and minutes",
			duration: 2*24*time.Hour + 5*time.Hour + 30*time.Minute + 15*time.Second,
			want:     "2d5h30m",
		},
		{
			name:     "exactly 1 day",
			duration: 24 * time.Hour,
			want:     "1d0h0m",
		},
		{
			name:     "days with no hours",
			duration: 3*24*time.Hour + 15*time.Minute,
			want:     "3d0h15m",
		},
		{
			name:     "hours and minutes",
			duration: 2*time.Hour + 30*time.Minute + 15*time.Second,
			want:     "2h30m",
		},
		{
			name:     "exactly 1 hour",
			duration: 1 * time.Hour,
			want:     "1h0m",
		},
		{
			name:     "minutes and seconds",
			duration: 5*time.Minute + 30*time.Second,
			want:     "5m30s",
		},
		{
			name:     "exactly 1 minute",
			duration: 1 * time.Minute,
			want:     "1m0s",
		},
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			want:     "45s",
		},
		{
			name:     "rounds to nearest second",
			duration: 1*time.Minute + 5*time.Second + 600*time.Millisecond,
			want:     "1m6s",
		},
		{
			name:     "rounds down",
			duration: 1*time.Minute + 5*time.Second + 400*time.Millisecond,
			want:     "1m5s",
		},
		{
			name:     "zero duration",
			duration: 0,
			want:     "0s",
		},
		{
			name:     "more than 1 day",
			duration: 25*time.Hour + 30*time.Minute,
			want:     "1d1h30m",
		},
		{
			name:     "hours no minutes",
			duration: 3*time.Hour + 10*time.Second,
			want:     "3h0m",
		},
		{
			name:     "7 days",
			duration: 7*24*time.Hour + 12*time.Hour + 45*time.Minute,
			want:     "7d12h45m",
		},
		{
			name:     "30 days",
			duration: 30*24*time.Hour + 6*time.Hour + 20*time.Minute,
			want:     "30d6h20m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUptime(tt.duration)
			assert.Equal(t, tt.want, got, "formatUptime(%v) should return %v", tt.duration, tt.want)
		})
	}
}
