package repository

import (
	"context"

	"github.com/haru-256/gcectl/internal/domain/model"
)

// VMRepository defines the interface for VM data access
type VMRepository interface {
	// FindByName retrieves a VM by its name, project, and zone
	FindByName(ctx context.Context, project, zone, name string) (*model.VM, error)

	// FindAll retrieves all VMs from the configuration
	FindAll(ctx context.Context) ([]*model.VM, error)

	// Start starts a VM instance
	Start(ctx context.Context, vm *model.VM) error

	// Stop stops a VM instance
	Stop(ctx context.Context, vm *model.VM) error

	// UpdateMachineType changes the machine type of a VM
	UpdateMachineType(ctx context.Context, vm *model.VM, machineType string) error

	// SetSchedulePolicy attaches a schedule policy to a VM
	SetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error

	// UnsetSchedulePolicy removes a schedule policy from a VM
	UnsetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error
}
