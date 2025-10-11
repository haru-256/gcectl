package model

import (
	"errors"
	"time"
)

// Status represents the operational state of a VM.
// It is an enumeration type that defines the possible states a GCE VM instance can be in.
type Status int

const (
	// StatusUnknown represents an unknown or unrecognized VM state
	StatusUnknown Status = iota
	// StatusRunning represents a VM that is currently running and ready to serve
	StatusRunning
	// StatusStopped represents a VM that has been stopped but can be restarted
	StatusStopped
	// StatusTerminated represents a VM that has been terminated
	StatusTerminated
	// StatusProvisioning represents a VM that is being created or started
	StatusProvisioning
)

// String returns the string representation of the VM status.
// This method implements the fmt.Stringer interface.
//
// Returns:
//   - "RUNNING" for StatusRunning
//   - "STOPPED" for StatusStopped
//   - "TERMINATED" for StatusTerminated
//   - "PROVISIONING" for StatusProvisioning
//   - "UNKNOWN" for StatusUnknown or any unrecognized status
func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "RUNNING"
	case StatusStopped:
		return "STOPPED"
	case StatusTerminated:
		return "TERMINATED"
	case StatusProvisioning:
		return "PROVISIONING"
	default:
		return "UNKNOWN"
	}
}

// StatusFromString converts a string representation to a Status type.
// This is useful for parsing status values from GCP API responses.
//
// Parameters:
//   - s: The status string to convert (e.g., "RUNNING", "STOPPED")
//
// Returns:
//   - The corresponding Status value, or StatusUnknown if the string is not recognized
func StatusFromString(s string) Status {
	switch s {
	case "RUNNING":
		return StatusRunning
	case "STOPPED":
		return StatusStopped
	case "TERMINATED":
		return StatusTerminated
	case "PROVISIONING":
		return StatusProvisioning
	default:
		return StatusUnknown
	}
}

// VM represents a Google Compute Engine virtual machine instance.
// This is the core domain model that encapsulates VM state and behavior.
// It is used throughout the application to represent VM instances consistently.
type VM struct {
	LastStartTime  *time.Time
	Name           string
	Project        string
	Zone           string
	MachineType    string
	SchedulePolicy string
	Status         Status
}

// Uptime calculates the current uptime of the VM if it is running.
//
// This method computes how long the VM has been running since its last start.
// It requires the VM to be in a RUNNING state and have a valid LastStartTime.
//
// Parameters:
//   - now: The current time to calculate uptime against
//
// Returns:
//   - time.Duration: The duration the VM has been running
//   - error: ErrVMNotRunning if the VM is not in RUNNING status,
//     ErrNoStartTime if LastStartTime is nil
//
// Example:
//
//	vm := &VM{Status: StatusRunning, LastStartTime: &startTime}
//	uptime, err := vm.Uptime(time.Now())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("VM has been running for %v\n", uptime)
func (v *VM) Uptime(now time.Time) (time.Duration, error) {
	if v.Status != StatusRunning {
		return 0, ErrVMNotRunning
	}
	if v.LastStartTime == nil {
		return 0, ErrNoStartTime
	}
	return now.Sub(*v.LastStartTime), nil
}

// CanStart checks if the VM can be started based on its current status.
//
// A VM can be started only if it is in STOPPED or TERMINATED status.
// This is a business rule that prevents attempting to start an already running VM.
//
// Returns:
//   - true if the VM is in STOPPED or TERMINATED status
//   - false otherwise (e.g., RUNNING, PROVISIONING, UNKNOWN)
func (v *VM) CanStart() bool {
	return v.Status == StatusStopped || v.Status == StatusTerminated
}

// CanStop checks if the VM can be stopped based on its current status.
//
// A VM can be stopped only if it is in RUNNING status.
// This is a business rule that prevents attempting to stop an already stopped VM.
//
// Returns:
//   - true if the VM is in RUNNING status
//   - false otherwise (e.g., STOPPED, TERMINATED, PROVISIONING, UNKNOWN)
func (v *VM) CanStop() bool {
	return v.Status == StatusRunning
}

var (
	ErrVMNotRunning = errors.New("VM is not running")
	ErrNoStartTime  = errors.New("VM start time is not available")
)
