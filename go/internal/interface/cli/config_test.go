package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	configPath := writeConfig(t, `default-project: test-project
default-zone: us-central1-a
vm:
  - name: test-vm
`)

	cfg, err := LoadConfig(configPath)

	require.NoError(t, err)
	require.Len(t, cfg.VMs, 1)
	assert.Equal(t, "test-project", cfg.VMs[0].Project)
	assert.Equal(t, "us-central1-a", cfg.VMs[0].Zone)
}

func TestResolveVMsByName(t *testing.T) {
	configPath := writeConfig(t, `default-project: test-project
default-zone: us-central1-a
vm:
  - name: vm-a
  - name: vm-b
    zone: asia-northeast1-a
`)

	vms, err := ResolveVMsByName(configPath, []string{"vm-b", "vm-a"})

	require.NoError(t, err)
	require.Len(t, vms, 2)
	assert.Equal(t, "vm-b", vms[0].Name)
	assert.Equal(t, "asia-northeast1-a", vms[0].Zone)
	assert.Equal(t, "vm-a", vms[1].Name)
	assert.Equal(t, "us-central1-a", vms[1].Zone)
}

func TestResolveVMsByNameReturnsMissingVM(t *testing.T) {
	configPath := writeConfig(t, `default-project: test-project
default-zone: us-central1-a
vm:
  - name: vm-a
`)

	vms, err := ResolveVMsByName(configPath, []string{"missing-vm"})

	require.Error(t, err)
	assert.Nil(t, vms)
	assert.Contains(t, err.Error(), "VM missing-vm not found")
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))
	return configPath
}
