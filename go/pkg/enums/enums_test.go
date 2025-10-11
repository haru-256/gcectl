package enums

import "testing"

func TestStatusFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Status
	}{
		{
			name:     "RUNNING status",
			input:    "RUNNING",
			expected: StatusRunning,
		},
		{
			name:     "TERMINATED status",
			input:    "TERMINATED",
			expected: StatusTerminated,
		},
		{
			name:     "Unknown status",
			input:    "STOPPED",
			expected: StatusUnknown,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: StatusUnknown,
		},
		{
			name:     "Lowercase status",
			input:    "running",
			expected: StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StatusFromString(tt.input)
			if result != tt.expected {
				t.Errorf("StatusFromString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		status   Status
	}{
		{
			name:     "Running status",
			status:   StatusRunning,
			expected: "RUNNING",
		},
		{
			name:     "Terminated status",
			status:   StatusTerminated,
			expected: "TERMINATED",
		},
		{
			name:     "Unknown status",
			status:   StatusUnknown,
			expected: "UNKNOWN",
		},
		{
			name:     "Invalid status value",
			status:   Status(99),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("Status.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStatus_Render(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		status   Status
	}{
		{
			name:     "Running status with emoji",
			status:   StatusRunning,
			expected: "🟢(RUNNING)",
		},
		{
			name:     "Terminated status with emoji",
			status:   StatusTerminated,
			expected: "🔴(TERMINATED)",
		},
		{
			name:     "Unknown status",
			status:   StatusUnknown,
			expected: "UNKNOWN",
		},
		{
			name:     "Invalid status value",
			status:   Status(99),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.Render()
			if result != tt.expected {
				t.Errorf("Status.Render() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStatusConstants(t *testing.T) {
	// Test that the constants have the expected values
	tests := []struct {
		name     string
		status   Status
		expected int
	}{
		{
			name:     "StatusUnknown is 0",
			status:   StatusUnknown,
			expected: 0,
		},
		{
			name:     "StatusRunning is 1",
			status:   StatusRunning,
			expected: 1,
		},
		{
			name:     "StatusTerminated is 2",
			status:   StatusTerminated,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.status) != tt.expected {
				t.Errorf("Status constant = %d, want %d", int(tt.status), tt.expected)
			}
		})
	}
}

func TestStatusRoundTrip(t *testing.T) {
	// Test that conversion from string to Status and back works correctly
	tests := []struct {
		name   string
		input  string
		output string
		status Status
	}{
		{
			name:   "RUNNING round trip",
			input:  "RUNNING",
			status: StatusRunning,
			output: "RUNNING",
		},
		{
			name:   "TERMINATED round trip",
			input:  "TERMINATED",
			status: StatusTerminated,
			output: "TERMINATED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// String -> Status
			status := StatusFromString(tt.input)
			if status != tt.status {
				t.Errorf("StatusFromString(%q) = %v, want %v", tt.input, status, tt.status)
			}

			// Status -> String
			output := status.String()
			if output != tt.output {
				t.Errorf("Status.String() = %q, want %q", output, tt.output)
			}
		})
	}
}
