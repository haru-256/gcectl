package config

import (
	"os"

	"github.com/haru-256/gcectl/pkg/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultProject string `yaml:"default-project"`
	DefaultZone    string `yaml:"default-zone"`
	VMs            []*VM  `yaml:"vm"`
}

type VM struct {
	Name           string
	Project        string
	Zone           string
	MachineType    string
	Status         string
	SchedulePolicy string
}

// GetVMByName returns the VM with the given name.
func (c *Config) GetVMByName(name string) *VM {
	for _, vm := range c.VMs {
		if vm.Name == name {
			return vm
		}
	}
	return nil
}

// ParseConfig parses the configuration file: confPath.
func ParseConfig(confPath string) (*Config, error) {
	cnf := Config{}

	data, err := os.ReadFile(confPath)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	err = yaml.Unmarshal(data, &cnf)
	if err != nil {
		return nil, err
	}

	return &cnf, nil
}
