package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		want        *Config
		configData  string
		configFile  string
		name        string
		errContains string
		wantErr     bool
	}{
		{
			name:       "valid config",
			configFile: "config.yaml",
			configData: `
default-project: test-project
default-zone: us-central1-a
vm:
  - name: instance-1
    project: test-project
    zone: us-central1-a
`,
			want: &Config{
				DefaultProject: "test-project",
				DefaultZone:    "us-central1-a",
				VMs: []*VM{
					{
						Name:    "instance-1",
						Project: "test-project",
						Zone:    "us-central1-a",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "invalid yaml",
			configFile: "invalid.yaml",
			configData: `invalid: yaml: [content`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:        "file not found",
			configFile:  "nonexistent.yaml",
			configData:  "",
			want:        nil,
			wantErr:     true,
			errContains: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tmpfile *os.File
			var err error

			if tt.configData != "" {
				// Create temporary config file
				tmpfile, err = os.CreateTemp("", tt.configFile)
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				defer os.Remove(tmpfile.Name())

				// Write test config data
				if err = os.WriteFile(tmpfile.Name(), []byte(tt.configData), 0644); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}

			// Test ParseConfig
			filename := "nonexistent.yaml"
			if tmpfile != nil {
				filename = tmpfile.Name()
			}

			got, err := ParseConfig(filename)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
