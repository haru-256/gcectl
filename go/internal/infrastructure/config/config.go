package config

import (
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

// ParseConfig reads a YAML configuration file and converts it to a Config structure.
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
//
// Example config.yaml:
//
//	default-project: my-project
//	default-zone: us-central1-a
//	vm:
//	  - name: vm1
//	    project: custom-project  # optional, uses default if omitted
//	    zone: us-west1-a         # optional, uses default if omitted
//	  - name: vm2                # will use default-project and default-zone
func ParseConfig(confPath string) (*Config, error) {
	data, err := os.ReadFile(confPath)
	if err != nil {
		// log パッケージは main や cmd など上位の層で利用するためここでは返却に留める
		return nil, err
	}

	// まずYAMLの構造を yamlConfig にデコードする
	var ymlCnf yamlConfig
	if unmarshalErr := yaml.Unmarshal(data, &ymlCnf); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// アプリケーションで利用する Config 構造体を作成
	cnf := &Config{
		DefaultProject: ymlCnf.DefaultProject,
		DefaultZone:    ymlCnf.DefaultZone,
	}

	// yamlVM のスライスから、ドメインモデルである model.VM のスライスへ変換する
	// このマッピング処理が、インフラ層の詳細とドメイン層を分離する重要な役割を果たします。
	for _, ymlVm := range ymlCnf.VMs {
		// VMごとのプロジェクトとゾーンが未指定の場合、デフォルト値を引き継ぐ
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
			// 他のフィールド (MachineType, Statusなど) は、
			// ユースケース層がリポジトリ経由で取得するため、ここでは初期化しない
		}
		cnf.VMs = append(cnf.VMs, vm)
	}

	return cnf, nil
}

// GetVMByName searches for a VM with the specified name in the configuration.
//
// This method searches through the configured VMs and returns the first VM
// that matches the given name. The comparison is case-sensitive.
//
// Parameters:
//   - name: The name of the VM to search for
//
// Returns:
//   - *model.VM: The VM with the matching name, or nil if not found
//
// Example:
//
//	vm := config.GetVMByName("my-vm")
//	if vm == nil {
//	    log.Fatal("VM not found")
//	}
func (c *Config) GetVMByName(name string) *model.VM {
	for _, vm := range c.VMs {
		if vm.Name == name {
			return vm
		}
	}
	return nil
}
