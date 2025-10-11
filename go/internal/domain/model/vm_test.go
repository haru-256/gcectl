package model

import (
	"testing"
	"time"
)

func TestStatus_String(t *testing.T) {
	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{
			name:   "running status",
			status: StatusRunning,
			want:   "RUNNING",
		},
		{
			name:   "stopped status",
			status: StatusStopped,
			want:   "STOPPED",
		},
		{
			name:   "terminated status",
			status: StatusTerminated,
			want:   "TERMINATED",
		},
		{
			name:   "provisioning status",
			status: StatusProvisioning,
			want:   "PROVISIONING",
		},
		{
			name:   "unknown status",
			status: StatusUnknown,
			want:   "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusFromString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Status
	}{
		{
			name:  "running string",
			input: "RUNNING",
			want:  StatusRunning,
		},
		{
			name:  "stopped string",
			input: "STOPPED",
			want:  StatusStopped,
		},
		{
			name:  "terminated string",
			input: "TERMINATED",
			want:  StatusTerminated,
		},
		{
			name:  "provisioning string",
			input: "PROVISIONING",
			want:  StatusProvisioning,
		},
		{
			name:  "unknown string",
			input: "INVALID",
			want:  StatusUnknown,
		},
		{
			name:  "empty string",
			input: "",
			want:  StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusFromString(tt.input); got != tt.want {
				t.Errorf("StatusFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVM_CanStart(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{
			name:   "can start when stopped",
			status: StatusStopped,
			want:   true,
		},
		{
			name:   "can start when terminated",
			status: StatusTerminated,
			want:   true,
		},
		{
			name:   "cannot start when running",
			status: StatusRunning,
			want:   false,
		},
		{
			name:   "cannot start when provisioning",
			status: StatusProvisioning,
			want:   false,
		},
		{
			name:   "cannot start when unknown",
			status: StatusUnknown,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{Status: tt.status}
			if got := vm.CanStart(); got != tt.want {
				t.Errorf("VM.CanStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVM_CanStop(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{
			name:   "can stop when running",
			status: StatusRunning,
			want:   true,
		},
		{
			name:   "cannot stop when stopped",
			status: StatusStopped,
			want:   false,
		},
		{
			name:   "cannot stop when terminated",
			status: StatusTerminated,
			want:   false,
		},
		{
			name:   "cannot stop when provisioning",
			status: StatusProvisioning,
			want:   false,
		},
		{
			name:   "cannot stop when unknown",
			status: StatusUnknown,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VM{Status: tt.status}
			if got := vm.CanStop(); got != tt.want {
				t.Errorf("VM.CanStop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVM_Uptime(t *testing.T) {
	startTime := time.Date(2025, 10, 11, 10, 0, 0, 0, time.UTC)
	now := time.Date(2025, 10, 11, 12, 30, 0, 0, time.UTC)
	expectedDuration := 2*time.Hour + 30*time.Minute

	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name         string
		vm           *VM
		now          time.Time
		wantDuration time.Duration
		wantErr      error
	}{
		{
			name: "success: running VM with start time",
			vm: &VM{
				Status:        StatusRunning,
				LastStartTime: &startTime,
			},
			now:          now,
			wantDuration: expectedDuration,
			wantErr:      nil,
		},
		{
			name: "error: VM not running",
			vm: &VM{
				Status:        StatusStopped,
				LastStartTime: &startTime,
			},
			now:          now,
			wantDuration: 0,
			wantErr:      ErrVMNotRunning,
		},
		{
			name: "error: no start time",
			vm: &VM{
				Status:        StatusRunning,
				LastStartTime: nil,
			},
			now:          now,
			wantDuration: 0,
			wantErr:      ErrNoStartTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := tt.vm.Uptime(tt.now)

			if err != tt.wantErr {
				t.Errorf("VM.Uptime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if duration != tt.wantDuration {
				t.Errorf("VM.Uptime() duration = %v, want %v", duration, tt.wantDuration)
			}
		})
	}
}
