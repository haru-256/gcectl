package config

import (
	"os"

	"github.com/haru-256/gce-commands/pkg/log"
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
	Status         string
	SchedulePolicy string
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
