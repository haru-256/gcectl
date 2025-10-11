package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/haru-256/gcectl/internal/domain/model"
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
				if cfg.DefaultProject != "test-project" {
					t.Errorf("DefaultProject = %v, want test-project", cfg.DefaultProject)
				}
				if cfg.DefaultZone != "us-central1-a" {
					t.Errorf("DefaultZone = %v, want us-central1-a", cfg.DefaultZone)
				}
				if len(cfg.VMs) != 2 {
					t.Fatalf("len(VMs) = %v, want 2", len(cfg.VMs))
				}
				if cfg.VMs[0].Name != "vm1" || cfg.VMs[0].Project != "project1" || cfg.VMs[0].Zone != "zone1" {
					t.Errorf("VM[0] = %+v, want {Name:vm1 Project:project1 Zone:zone1}", cfg.VMs[0])
				}
				if cfg.VMs[1].Name != "vm2" || cfg.VMs[1].Project != "project2" || cfg.VMs[1].Zone != "zone2" {
					t.Errorf("VM[1] = %+v, want {Name:vm2 Project:project2 Zone:zone2}", cfg.VMs[1])
				}
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
				if len(cfg.VMs) != 3 {
					t.Fatalf("len(VMs) = %v, want 3", len(cfg.VMs))
				}
				// vm1 should inherit both defaults
				if cfg.VMs[0].Project != "default-proj" || cfg.VMs[0].Zone != "default-zone" {
					t.Errorf("VM[0] project/zone = %v/%v, want default-proj/default-zone", cfg.VMs[0].Project, cfg.VMs[0].Zone)
				}
				// vm2 has custom project, inherits default zone
				if cfg.VMs[1].Project != "custom-proj" || cfg.VMs[1].Zone != "default-zone" {
					t.Errorf("VM[1] project/zone = %v/%v, want custom-proj/default-zone", cfg.VMs[1].Project, cfg.VMs[1].Zone)
				}
				// vm3 has custom zone, inherits default project
				if cfg.VMs[2].Project != "default-proj" || cfg.VMs[2].Zone != "custom-zone" {
					t.Errorf("VM[2] project/zone = %v/%v, want default-proj/custom-zone", cfg.VMs[2].Project, cfg.VMs[2].Zone)
				}
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
					t.Fatalf("Failed to create temp config: %v", writeErr)
				}
			}

			cfg, err := ParseConfig(confPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseConfig() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

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
				if vm != nil {
					t.Errorf("GetVMByName() = %v, want nil", vm)
				}
				return
			}

			if vm == nil {
				t.Errorf("GetVMByName() = nil, want %v", tt.wantVM)
				return
			}

			if vm.Name != tt.wantVM.Name || vm.Project != tt.wantVM.Project || vm.Zone != tt.wantVM.Zone {
				t.Errorf("GetVMByName() = %+v, want %+v", vm, tt.wantVM)
			}
		})
	}
}
