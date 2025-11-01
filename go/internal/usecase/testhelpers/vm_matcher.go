package testhelpers

import (
	"context"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

// AssertVMInput validates that the input VM has the expected Project, Zone, and Name.
// This is commonly used in DoAndReturn to verify FindByName arguments.
func AssertVMInput(t *testing.T, inputVM *model.VM, expectedProject, expectedZone, expectedName string) {
	t.Helper()
	assert.Equal(t, expectedProject, inputVM.Project, "VM.Project should match")
	assert.Equal(t, expectedZone, inputVM.Zone, "VM.Zone should match")
	assert.Equal(t, expectedName, inputVM.Name, "VM.Name should match")
}

// VMFindByNameMatcher creates a DoAndReturn function for FindByName that validates input and returns the specified VM.
func VMFindByNameMatcher(t *testing.T, expectedVM *model.VM, returnVM *model.VM, returnErr error) func(context.Context, *model.VM) (*model.VM, error) {
	t.Helper()
	return func(ctx context.Context, inputVM *model.VM) (*model.VM, error) {
		AssertVMInput(t, inputVM, expectedVM.Project, expectedVM.Zone, expectedVM.Name)
		return returnVM, returnErr
	}
}
