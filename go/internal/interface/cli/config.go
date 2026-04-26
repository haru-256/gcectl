package cli

import (
	"fmt"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/infrastructure/config"
)

// LoadConfig reads the gcectl configuration from disk.
func LoadConfig(configPath string) (*config.Config, error) {
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// ResolveVMsByName loads configured VMs in the same order requested by name.
func ResolveVMsByName(configPath string, names []string) ([]*model.VM, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	vms := make([]*model.VM, 0, len(names))
	for _, name := range names {
		vm := cfg.GetVMByName(name)
		if vm == nil {
			return nil, fmt.Errorf("VM %s not found", name)
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

// ResolveVMByName loads one configured VM by name.
func ResolveVMByName(configPath, name string) (*model.VM, error) {
	vms, err := ResolveVMsByName(configPath, []string{name})
	if err != nil {
		return nil, err
	}
	return vms[0], nil
}
