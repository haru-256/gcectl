package presenter

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
)

func TestNewConsolePresenter(t *testing.T) {
	presenter := NewConsolePresenter()

	if presenter == nil {
		t.Fatal("NewConsolePresenter() returned nil")
	}
}

func TestConsolePresenter_Success(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.Success("Test success message")

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[SUCCESS]") {
		t.Errorf("Success() output = %q, want to contain [SUCCESS]", output)
	}

	if !strings.Contains(output, "Test success message") {
		t.Errorf("Success() output = %q, want to contain 'Test success message'", output)
	}
}

func TestConsolePresenter_Error(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.Error("Test error message")

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Error() output = %q, want to contain [ERROR]", output)
	}

	if !strings.Contains(output, "Test error message") {
		t.Errorf("Error() output = %q, want to contain 'Test error message'", output)
	}
}

func TestConsolePresenter_Progress(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.Progress()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if output != "." {
		t.Errorf("Progress() output = %q, want '.'", output)
	}
}

func TestConsolePresenter_ProgressDone(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.ProgressDone()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if output != "\n" {
		t.Errorf("ProgressDone() output = %q, want newline", output)
	}
}

func TestConsolePresenter_ProgressSequence(t *testing.T) {
	presenter := NewConsolePresenter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Simulate a sequence of progress updates
	presenter.Progress()
	presenter.Progress()
	presenter.Progress()
	presenter.ProgressDone()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	expected := "...\n"
	if output != expected {
		t.Errorf("Progress sequence output = %q, want %q", output, expected)
	}
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
			Uptime:         "2h30m15s",
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
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.RenderVMList(items)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Check that VM names appear in output
	if !strings.Contains(output, "vm1") {
		t.Errorf("RenderVMList() output doesn't contain 'vm1'")
	}

	if !strings.Contains(output, "vm2") {
		t.Errorf("RenderVMList() output doesn't contain 'vm2'")
	}

	// Check that table headers appear
	expectedHeaders := []string{"Name", "Project", "Zone", "Machine-Type", "Status"}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("RenderVMList() output doesn't contain header '%s'", header)
		}
	}

	// Check that uptime values appear
	if !strings.Contains(output, "2h30m15s") {
		t.Errorf("RenderVMList() output doesn't contain uptime '2h30m15s'")
	}

	if !strings.Contains(output, "N/A") {
		t.Errorf("RenderVMList() output doesn't contain uptime 'N/A'")
	}
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
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.RenderVMDetail(detail)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
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
		if !strings.Contains(output, field) {
			t.Errorf("RenderVMDetail() output doesn't contain '%s'", field)
		}
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
			if got != tt.want {
				t.Errorf("getStatusEmoji(%v) = %v, want %v", tt.status, got, tt.want)
			}
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

			if len(paddings) != tt.wantLen {
				t.Errorf("getItemPaddings() returned %d paddings, want %d", len(paddings), tt.wantLen)
			}

			// Check that paddings are strings (possibly empty)
			for i, padding := range paddings {
				if padding != "" && !strings.HasPrefix(padding+"x", " ") {
					t.Errorf("padding[%d] = %q is not a valid padding string", i, padding)
				}
			}
		})
	}
}
