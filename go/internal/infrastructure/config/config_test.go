package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gocognit // cognitive complexity is less important than readability in tests
func TestParseConfig(t *testing.T) {
	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name         string
		yamlContent  string
		wantErr      bool
		validateFunc func(*testing.T, *Config)
	}{
		{
			name: "success: valid config with all fields",
			yamlContent: `default-project: test-project
default-zone: us-central1-a
vm:
  - name: vm1
    project: project1
    zone: zone1
  - name: vm2
    project: project2
    zone: zone2
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "test-project", cfg.DefaultProject, "DefaultProject should be test-project")
				assert.Equal(t, "us-central1-a", cfg.DefaultZone, "DefaultZone should be us-central1-a")
				require.Len(t, cfg.VMs, 2, "VMs should have 2 entries")
				assert.Equal(t, "vm1", cfg.VMs[0].Name, "VM[0].Name should be vm1")
				assert.Equal(t, "project1", cfg.VMs[0].Project, "VM[0].Project should be project1")
				assert.Equal(t, "zone1", cfg.VMs[0].Zone, "VM[0].Zone should be zone1")
				assert.Equal(t, "vm2", cfg.VMs[1].Name, "VM[1].Name should be vm2")
				assert.Equal(t, "project2", cfg.VMs[1].Project, "VM[1].Project should be project2")
				assert.Equal(t, "zone2", cfg.VMs[1].Zone, "VM[1].Zone should be zone2")
			},
		},
		{
			name: "success: VMs inherit default project and zone",
			yamlContent: `default-project: default-proj
default-zone: default-zone
vm:
  - name: vm1
  - name: vm2
    project: custom-proj
  - name: vm3
    zone: custom-zone
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *Config) {
				require.Len(t, cfg.VMs, 3, "VMs should have 3 entries")
				// vm1 should inherit both defaults
				assert.Equal(t, "default-proj", cfg.VMs[0].Project, "VM[0].Project should be default-proj")
				assert.Equal(t, "default-zone", cfg.VMs[0].Zone, "VM[0].Zone should be default-zone")
				// vm2 has custom project, inherits default zone
				assert.Equal(t, "custom-proj", cfg.VMs[1].Project, "VM[1].Project should be custom-proj")
				assert.Equal(t, "default-zone", cfg.VMs[1].Zone, "VM[1].Zone should be default-zone")
				// vm3 has custom zone, inherits default project
				assert.Equal(t, "default-proj", cfg.VMs[2].Project, "VM[2].Project should be default-proj")
				assert.Equal(t, "custom-zone", cfg.VMs[2].Zone, "VM[2].Zone should be custom-zone")
			},
		},
		{
			name:         "error: file not found",
			yamlContent:  "",
			wantErr:      true,
			validateFunc: nil,
		},
		{
			name: "error: invalid YAML",
			yamlContent: `default-project: test
invalid yaml syntax: [
`,
			wantErr:      true,
			validateFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var confPath string

			if tt.name == "error: file not found" {
				confPath = "/nonexistent/path/config.yaml"
			} else {
				// Create temporary config file
				tmpDir := t.TempDir()
				confPath = filepath.Join(tmpDir, "config.yaml")

				if writeErr := os.WriteFile(confPath, []byte(tt.yamlContent), 0644); writeErr != nil {
					require.NoError(t, writeErr, "Failed to create temp config")
				}
			}

			cfg, err := ParseConfig(confPath)

			if tt.wantErr {
				assert.Error(t, err, "ParseConfig() should return an error")
				return
			}

			assert.NoError(t, err, "ParseConfig() should not return an error")

			if tt.validateFunc != nil {
				tt.validateFunc(t, cfg)
			}
		})
	}
}

func TestConfig_GetVMByName(t *testing.T) {
	cfg := &Config{
		DefaultProject: "test-project",
		DefaultZone:    "us-central1-a",
		VMs: []*model.VM{
			{Name: "vm1", Project: "project1", Zone: "zone1"},
			{Name: "vm2", Project: "project2", Zone: "zone2"},
			{Name: "vm3", Project: "project3", Zone: "zone3"},
		},
	}

	//nolint:govet // field alignment is less important than readability in tests
	tests := []struct {
		name    string
		vmName  string
		wantVM  *model.VM
		wantNil bool
	}{
		{
			name:    "success: find existing VM",
			vmName:  "vm2",
			wantVM:  &model.VM{Name: "vm2", Project: "project2", Zone: "zone2"},
			wantNil: false,
		},
		{
			name:    "success: find first VM",
			vmName:  "vm1",
			wantVM:  &model.VM{Name: "vm1", Project: "project1", Zone: "zone1"},
			wantNil: false,
		},
		{
			name:    "not found: nonexistent VM",
			vmName:  "nonexistent",
			wantVM:  nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := cfg.GetVMByName(tt.vmName)

			if tt.wantNil {
				assert.Nil(t, vm, "GetVMByName() should return nil")
				return
			}

			require.NotNil(t, vm, "GetVMByName() should not return nil")
			assert.Equal(t, tt.wantVM.Name, vm.Name, "VM.Name should match")
			assert.Equal(t, tt.wantVM.Project, vm.Project, "VM.Project should match")
			assert.Equal(t, tt.wantVM.Zone, vm.Zone, "VM.Zone should match")
		})
	}
}
