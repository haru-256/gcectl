package presenter

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

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

func TestConsolePresenter_RenderVMList(t *testing.T) {
	presenter := NewConsolePresenter()
	startTime := time.Date(2025, 10, 11, 10, 0, 0, 0, time.UTC)

	vms := []*model.VM{
		{
			Name:           "vm1",
			Project:        "project1",
			Zone:           "us-central1-a",
			MachineType:    "e2-medium",
			Status:         model.StatusRunning,
			SchedulePolicy: "policy1",
			LastStartTime:  &startTime,
		},
		{
			Name:           "vm2",
			Project:        "project2",
			Zone:           "us-west1-a",
			MachineType:    "n1-standard-1",
			Status:         model.StatusStopped,
			SchedulePolicy: "",
			LastStartTime:  nil,
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.RenderVMList(vms)

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
}

func TestConsolePresenter_RenderVMDetail(t *testing.T) {
	presenter := NewConsolePresenter()

	vm := &model.VM{
		Name:           "test-vm",
		Project:        "test-project",
		Zone:           "us-central1-a",
		MachineType:    "e2-medium",
		Status:         model.StatusRunning,
		SchedulePolicy: "test-policy",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	presenter.RenderVMDetail(vm)

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
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("RenderVMDetail() output doesn't contain '%s'", field)
		}
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
