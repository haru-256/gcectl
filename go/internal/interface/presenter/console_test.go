package presenter

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsolePresenter(t *testing.T) {
	presenter := NewConsolePresenter()

	require.NotNil(t, presenter, "NewConsolePresenter() should not return nil")
	assert.NotNil(t, presenter.errorStyle, "errorStyle should be initialized")
	assert.NotNil(t, presenter.successStyle, "successStyle should be initialized")
}

func TestConsolePresenter_Success(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	presenter.Success("Test success message")

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	assert.Contains(t, output, "[SUCCESS]", "Output should contain [SUCCESS]")
	assert.Contains(t, output, "Test success message", "Output should contain the test message")
}

func TestConsolePresenter_Error(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	presenter.Error("Test error message")

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	assert.Contains(t, output, "[ERROR]", "Output should contain [ERROR]")
	assert.Contains(t, output, "Test error message", "Output should contain the test message")
}

func TestConsolePresenter_Progress(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	presenter.Progress()

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	assert.Equal(t, ".", output, "Progress() should output a single dot")
}

func TestConsolePresenter_ProgressDone(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	presenter.ProgressDone()

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	assert.Equal(t, "\n", output, "ProgressDone() should output a newline")
}

func TestConsolePresenter_ProgressStart(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	message := "Starting VM test-vm"
	presenter.ProgressStart(message)

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	assert.Equal(t, message, output, "ProgressStart() should output the provided message")
}

func TestConsolePresenter_ProgressSequence(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	// Simulate a sequence of progress updates
	presenter.Progress()
	presenter.Progress()
	presenter.Progress()
	presenter.ProgressDone()

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	assert.Equal(t, "...\n", output, "Progress sequence should output dots followed by newline")
}

func TestConsolePresenter_ProgressStartWithSequence(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	// Simulate a complete progress sequence with start message
	presenter.ProgressStart("Starting VM test-vm")
	presenter.Progress()
	presenter.Progress()
	presenter.ProgressDone()

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	assert.Equal(t, "Starting VM test-vm..\n", output, "Complete progress sequence should show message, dots, and newline")
}

func TestConsolePresenter_RenderVMList(t *testing.T) {
	presenter := NewConsolePresenter()

	items := []VMListItem{
		{
			Name:           "vm1",
			Project:        "project1",
			Zone:           "us-central1-a",
			MachineType:    "e2-medium",
			Status:         model.StatusRunning,
			SchedulePolicy: "policy1",
			Uptime:         "2h30m",
		},
		{
			Name:           "vm2",
			Project:        "project2",
			Zone:           "us-west1-a",
			MachineType:    "n1-standard-1",
			Status:         model.StatusStopped,
			SchedulePolicy: "",
			Uptime:         "N/A",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	presenter.RenderVMList(items)

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	// Check that VM names appear in output
	assert.Contains(t, output, "vm1", "Output should contain vm1")
	assert.Contains(t, output, "vm2", "Output should contain vm2")

	// Check that table headers appear
	expectedHeaders := []string{"Name", "Project", "Zone", "Machine-Type", "Status"}
	for _, header := range expectedHeaders {
		assert.Contains(t, output, header, "Output should contain header '%s'", header)
	}

	// Check that uptime values appear
	assert.Contains(t, output, "2h30m", "Output should contain uptime '2h30m'")
	assert.Contains(t, output, "N/A", "Output should contain uptime 'N/A'")
}

func TestConsolePresenter_RenderVMDetail(t *testing.T) {
	presenter := NewConsolePresenter()

	detail := VMDetail{
		Name:           "test-vm",
		Project:        "test-project",
		Zone:           "us-central1-a",
		MachineType:    "e2-medium",
		Status:         model.StatusRunning,
		SchedulePolicy: "test-policy",
		Uptime:         "2h30m",
	}

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")
	os.Stdout = w

	presenter.RenderVMDetail(detail)

	require.NoError(t, w.Close(), "Failed to close write pipe")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy output")
	output := buf.String()

	// Check that VM details appear in output
	expectedFields := []string{
		"test-vm",
		"test-project",
		"us-central1-a",
		"e2-medium",
		"RUNNING",
		"test-policy",
		"2h30m",
	}

	for _, field := range expectedFields {
		assert.Contains(t, output, field, "Output should contain field '%s'", field)
	}
}

func TestGetStatusEmoji(t *testing.T) {
	//nolint:govet // Test struct prioritizes readability over field alignment
	tests := []struct {
		name   string
		status model.Status
		want   string
	}{
		{
			name:   "running status",
			status: model.StatusRunning,
			want:   "🟢",
		},
		{
			name:   "stopped status",
			status: model.StatusStopped,
			want:   "🔴",
		},
		{
			name:   "terminated status",
			status: model.StatusTerminated,
			want:   "🔴",
		},
		{
			name:   "provisioning status",
			status: model.StatusProvisioning,
			want:   "⚪",
		},
		{
			name:   "unknown status",
			status: model.StatusUnknown,
			want:   "⚪",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStatusEmoji(tt.status)
			assert.Equal(t, tt.want, got, "getStatusEmoji(%v) should return %v", tt.status, tt.want)
		})
	}
}

func TestGetItemPaddings(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		wantLen int
	}{
		{
			name:    "equal length headers",
			headers: []string{"Name", "Zone", "Type"},
			wantLen: 3,
		},
		{
			name:    "varying length headers",
			headers: []string{"Name", "Project", "Zone", "MachineType"},
			wantLen: 4,
		},
		{
			name:    "single header",
			headers: []string{"Name"},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paddings := getItemPaddings(tt.headers)

			assert.Len(t, paddings, tt.wantLen, "getItemPaddings() should return %d paddings", tt.wantLen)

			// Check that paddings are strings (possibly empty)
			for i, padding := range paddings {
				if padding != "" {
					assert.True(t, len(padding) > 0 && padding[0] == ' ', "padding[%d] = %q should start with space or be empty", i, padding)
				}
			}
		})
	}
}
