//go:build integration || !ci

package gcp_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/infrastructure/config"
	"github.com/haru-256/gcectl/internal/infrastructure/gcp"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logger = log.NewLogger()

func getCnf(t *testing.T) (*config.Config, string) {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	cnfPath := home + "/.config/gcectl/config.yaml"
	cnf, err := config.ParseConfig(cnfPath)
	require.NoError(t, err)
	require.NotNil(t, cnf)
	require.NotEqual(t, "", cnf.DefaultProject)
	require.NotEqual(t, "", cnf.DefaultZone)
	require.GreaterOrEqual(t, len(cnf.VMs), 2)
	return cnf, cnfPath
}

func TestVMRepositoryImpl_FindByName(t *testing.T) {
	cnf, cnfPath := getCnf(t)
	repo := gcp.NewVMRepository(cnfPath, logger)
	ctx := context.Background()

	tests := []struct {
		name      string
		vm        *model.VM
		expectErr bool
	}{
		{
			name: "existing VM",
			vm: &model.VM{
				Project: cnf.VMs[0].Project,
				Zone:    cnf.VMs[0].Zone,
				Name:    cnf.VMs[0].Name,
			},
			expectErr: false,
		},
		{
			name: "not existing VM",
			vm: &model.VM{
				Project: "hoge-project",
				Zone:    "us-central1-a",
				Name:    "hoge-vm",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundVM, err := repo.FindByName(ctx, tt.vm)
			if tt.expectErr {
				require.Error(t, err)
				require.Nil(t, foundVM)
			} else {
				require.NoError(t, err)
				require.NotNil(t, foundVM)
				require.Equal(t, tt.vm.Project, foundVM.Project)
				require.Equal(t, tt.vm.Zone, foundVM.Zone)
				require.Equal(t, tt.vm.Name, foundVM.Name)
				// Verify required fields are populated
				assert.NotEmpty(t, foundVM.MachineType)
				assert.NotEqual(t, model.StatusUnknown, foundVM.Status)
			}
		})
	}
}

func TestVMRepositoryImpl_FindAll(t *testing.T) {
	cnf, cnfPath := getCnf(t)
	repo := gcp.NewVMRepository(cnfPath, logger)
	ctx := context.Background()

	vms, err := repo.FindAll(ctx)
	require.NoError(t, err)
	require.NotNil(t, vms)
	require.GreaterOrEqual(t, len(vms), len(cnf.VMs), "Should find at least the configured VMs")

	// Verify that all configured VMs are found
	for _, configVM := range cnf.VMs {
		found := false
		for _, vm := range vms {
			if vm.Project == configVM.Project && vm.Zone == configVM.Zone && vm.Name == configVM.Name {
				found = true
				// Verify required fields are populated
				assert.NotEmpty(t, vm.MachineType)
				assert.NotEqual(t, model.StatusUnknown, vm.Status)
				break
			}
		}
		assert.True(t, found, "Configured VM %s/%s/%s should be found", configVM.Project, configVM.Zone, configVM.Name)
	}
}

func TestVMRepositoryImpl_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running integration test in short mode")
	}

	cnf, cnfPath := getCnf(t)
	repo := gcp.NewVMRepository(cnfPath, logger)
	ctx := context.Background()

	// Use the first configured VM for testing
	testVM := &model.VM{
		Project: cnf.VMs[0].Project,
		Zone:    cnf.VMs[0].Zone,
		Name:    cnf.VMs[0].Name,
	}

	// Find the current state
	currentVM, err := repo.FindByName(ctx, testVM)
	require.NoError(t, err)
	require.NotNil(t, currentVM)

	t.Run("stop and start VM", func(t *testing.T) {
		// If VM is running, stop it first
		if currentVM.Status == model.StatusRunning {
			t.Log("Stopping running VM...")
			stopErr := repo.Stop(ctx, currentVM)
			require.NoError(t, stopErr)

			// Wait for VM to be fully stopped
			require.Eventually(t, func() bool {
				vm, findErr := repo.FindByName(ctx, testVM)
				if findErr != nil {
					return false
				}
				return vm.Status == model.StatusStopped || vm.Status == model.StatusTerminated
			}, 2*time.Minute, 5*time.Second, "VM should be stopped")
		}

		// Start the VM
		t.Log("Starting VM...")
		stoppedVM, findErr := repo.FindByName(ctx, testVM)
		require.NoError(t, findErr)
		startErr := repo.Start(ctx, stoppedVM)
		require.NoError(t, startErr)

		// Wait for VM to be running
		require.Eventually(t, func() bool {
			vm, checkErr := repo.FindByName(ctx, testVM)
			if checkErr != nil {
				return false
			}
			return vm.Status == model.StatusRunning
		}, 2*time.Minute, 5*time.Second, "VM should be running")

		// Verify VM is running
		runningVM, findErr := repo.FindByName(ctx, testVM)
		require.NoError(t, findErr)
		assert.Equal(t, model.StatusRunning, runningVM.Status)
		assert.NotNil(t, runningVM.LastStartTime)

		// Stop the VM again
		t.Log("Stopping VM again...")
		stopErr := repo.Stop(ctx, runningVM)
		require.NoError(t, stopErr)

		// Wait for VM to be stopped
		require.Eventually(t, func() bool {
			vm, checkErr := repo.FindByName(ctx, testVM)
			if checkErr != nil {
				return false
			}
			return vm.Status == model.StatusStopped || vm.Status == model.StatusTerminated
		}, 2*time.Minute, 5*time.Second, "VM should be stopped")

		// Verify VM is stopped
		finalVM, findErr := repo.FindByName(ctx, testVM)
		require.NoError(t, findErr)
		assert.True(t, finalVM.Status == model.StatusStopped || finalVM.Status == model.StatusTerminated)
	})
}

func TestVMRepositoryImpl_UpdateMachineType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running integration test in short mode")
	}

	cnf, cnfPath := getCnf(t)
	repo := gcp.NewVMRepository(cnfPath, logger)
	ctx := context.Background()

	// Use the first configured VM for testing
	testVM := &model.VM{
		Project: cnf.VMs[0].Project,
		Zone:    cnf.VMs[0].Zone,
		Name:    cnf.VMs[0].Name,
	}

	// Find the current state and machine type
	currentVM, err := repo.FindByName(ctx, testVM)
	require.NoError(t, err)
	require.NotNil(t, currentVM)
	originalMachineType := currentVM.MachineType

	// Ensure VM is stopped
	if currentVM.Status == model.StatusRunning {
		t.Log("Stopping VM for machine type update...")
		stopErr := repo.Stop(ctx, currentVM)
		require.NoError(t, stopErr)

		require.Eventually(t, func() bool {
			vm, findErr := repo.FindByName(ctx, testVM)
			if findErr != nil {
				return false
			}
			return vm.Status == model.StatusStopped || vm.Status == model.StatusTerminated
		}, 2*time.Minute, 5*time.Second, "VM should be stopped")
	}

	t.Run("update machine type", func(t *testing.T) {
		stoppedVM, findErr := repo.FindByName(ctx, testVM)
		require.NoError(t, findErr)

		// Choose a different machine type
		newMachineType := "e2-micro"
		if originalMachineType == "e2-micro" {
			newMachineType = "e2-small"
		}

		t.Logf("Updating machine type from %s to %s...", originalMachineType, newMachineType)
		updateErr := repo.UpdateMachineType(ctx, stoppedVM, newMachineType)
		require.NoError(t, updateErr)

		// Wait for update to complete
		require.Eventually(t, func() bool {
			vm, checkErr := repo.FindByName(ctx, testVM)
			if checkErr != nil {
				return false
			}
			return vm.MachineType == newMachineType
		}, 2*time.Minute, 5*time.Second, "Machine type should be updated")

		// Verify machine type was updated
		updatedVM, findErr := repo.FindByName(ctx, testVM)
		require.NoError(t, findErr)
		assert.Equal(t, newMachineType, updatedVM.MachineType)

		// Restore original machine type
		t.Logf("Restoring machine type to %s...", originalMachineType)
		restoreErr := repo.UpdateMachineType(ctx, updatedVM, originalMachineType)
		require.NoError(t, restoreErr)

		require.Eventually(t, func() bool {
			vm, checkErr := repo.FindByName(ctx, testVM)
			if checkErr != nil {
				return false
			}
			return vm.MachineType == originalMachineType
		}, 2*time.Minute, 5*time.Second, "Machine type should be restored")
	})
}

func TestVMRepositoryImpl_SchedulePolicy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running integration test in short mode")
	}

	cnf, cnfPath := getCnf(t)
	repo := gcp.NewVMRepository(cnfPath, logger)
	ctx := context.Background()

	// Use the first configured VM for testing
	testVM := &model.VM{
		Project: cnf.VMs[0].Project,
		Zone:    cnf.VMs[0].Zone,
		Name:    cnf.VMs[0].Name,
	}

	// Find the current state
	currentVM, err := repo.FindByName(ctx, testVM)
	require.NoError(t, err)
	require.NotNil(t, currentVM)
	originalPolicy := currentVM.SchedulePolicy

	// Test policy name - should exist in the project
	testPolicyName := "test-schedule-policy"

	t.Run("set and unset schedule policy", func(t *testing.T) {
		// Set schedule policy
		t.Logf("Setting schedule policy to %s...", testPolicyName)
		setErr := repo.SetSchedulePolicy(ctx, currentVM, testPolicyName)
		if setErr != nil {
			// If the test policy doesn't exist, skip this test
			t.Skipf("Schedule policy '%s' not found in project, skipping test: %v", testPolicyName, setErr)
		}

		// Wait for policy to be set
		require.Eventually(t, func() bool {
			vm, checkErr := repo.FindByName(ctx, testVM)
			if checkErr != nil {
				return false
			}
			return vm.SchedulePolicy == testPolicyName
		}, 2*time.Minute, 5*time.Second, "Schedule policy should be set")

		// Verify policy was set
		updatedVM, findErr := repo.FindByName(ctx, testVM)
		require.NoError(t, findErr)
		assert.Equal(t, testPolicyName, updatedVM.SchedulePolicy)

		// Unset schedule policy
		t.Log("Unsetting schedule policy...")
		unsetErr := repo.UnsetSchedulePolicy(ctx, updatedVM, testPolicyName)
		require.NoError(t, unsetErr)

		// Wait for policy to be unset
		require.Eventually(t, func() bool {
			vm, checkErr := repo.FindByName(ctx, testVM)
			if checkErr != nil {
				return false
			}
			return vm.SchedulePolicy == ""
		}, 2*time.Minute, 5*time.Second, "Schedule policy should be unset")

		// Verify policy was unset
		finalVM, findErr := repo.FindByName(ctx, testVM)
		require.NoError(t, findErr)
		assert.Empty(t, finalVM.SchedulePolicy)

		// Restore original policy if it existed
		if originalPolicy != "" {
			t.Logf("Restoring original schedule policy %s...", originalPolicy)
			restoreErr := repo.SetSchedulePolicy(ctx, finalVM, originalPolicy)
			require.NoError(t, restoreErr)
		}
	})
}
