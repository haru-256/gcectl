package config

import (
	"fmt"
	"os"

	"github.com/haru-256/gcectl/internal/domain/model" // ドメインモデルをインポート
	"gopkg.in/yaml.v3"
)

// Config holds the application-wide configuration settings.
// It maintains a list of VMs as domain models and provides access methods.
// This structure abstracts away the underlying YAML file format from the rest of the application.
type Config struct {
	DefaultProject string
	DefaultZone    string
	VMs            []*model.VM // ドメインモデルのVMを参照
}

// yamlConfig is a temporary structure that directly maps the config.yaml file format.
// This structure is used only within this package for unmarshaling YAML content.
type yamlConfig struct {
	DefaultProject string   `yaml:"default-project"`
	DefaultZone    string   `yaml:"default-zone"`
	VMs            []yamlVM `yaml:"vm"`
}

// yamlVM is a temporary structure that maps a VM entry in config.yaml.
// This structure is used only within this package for unmarshaling YAML content.
type yamlVM struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
	Zone    string `yaml:"zone"`
}

// NewConfig reads a YAML configuration file and converts it to a Config structure.
//
// This function performs the following steps:
// 1. Reads the YAML file from the specified path
// 2. Unmarshals the YAML content into a yamlConfig structure
// 3. Converts yamlConfig to Config with domain model VMs
// 4. Applies default project/zone to VMs that don't specify them
//
// Parameters:
//   - confPath: The file path to the YAML configuration file
//
// Returns:
//   - *Config: The parsed configuration with domain model VMs
//   - error: An error if file reading or YAML parsing fails
func NewConfig(confPath string) (*Config, error) {
	data, err := os.ReadFile(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var ymlCnf yamlConfig
	if unmarshalErr := yaml.Unmarshal(data, &ymlCnf); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", unmarshalErr)
	}

	cnf := &Config{
		DefaultProject: ymlCnf.DefaultProject,
		DefaultZone:    ymlCnf.DefaultZone,
	}

	for _, ymlVm := range ymlCnf.VMs {
		project := ymlVm.Project
		if project == "" {
			project = ymlCnf.DefaultProject
		}
		zone := ymlVm.Zone
		if zone == "" {
			zone = ymlCnf.DefaultZone
		}

		vm := &model.VM{
			Name:    ymlVm.Name,
			Project: project,
			Zone:    zone,
		}
		cnf.VMs = append(cnf.VMs, vm)
	}

	return cnf, nil
}

// getVMByName searches for a VM with the specified name in the configuration.
func (c *Config) getVMByName(name string) *model.VM {
	for _, vm := range c.VMs {
		if vm.Name == name {
			return vm
		}
	}
	return nil
}

// ResolveVMs returns VM domain models matching the given names.
// It maintains the order of names requested.
func (c *Config) ResolveVMs(names []string) ([]*model.VM, error) {
	vms := make([]*model.VM, 0, len(names))
	for _, name := range names {
		vm := c.getVMByName(name)
		if vm == nil {
			return nil, fmt.Errorf("VM %s not found in config", name)
		}
		vms = append(vms, vm)
	}
	return vms, nil
}

// ResolveVM returns a single VM domain model matching the given name.
func (c *Config) ResolveVM(name string) (*model.VM, error) {
	vm := c.getVMByName(name)
	if vm == nil {
		return nil, fmt.Errorf("VM %s not found in config", name)
	}
	return vm, nil
}
