# gcectl: Google Cloud Compute Engine Commands

A CLI tool to simplify management of Google Cloud Compute Engine instances. This tool provides convenient shortcuts for common GCE operations like starting/stopping VMs, changing machine types, and setting scheduling policies.

## Installation

### From Source

```bash
git clone https://github.com/haru-256/gcectl.git
cd gcectl/golang
go install
```

## Prerequisites

- Go 1.21 or higher (for building from source)
- Google Cloud SDK installed and configured
- Active Google Cloud Project with Compute Engine API enabled
- Proper permissions to manage Compute Engine resources

## Usage

```sh
# Describe a VM
gcectl describe <vm_name>

# Turn on a VM
gcectl on <vm_name>

# Turn off a VM
gcectl off <vm_name>

# Set machine type
gcectl set machine-type <vm_name> <machine-type>
# Example: gcectl set machine-type sandbox n1-standard-1

# Set schedule policy
gcectl set schedule-policy <vm_name> <policy_name>
# Example: gcectl set schedule-policy sandbox stop

# Unset schedule policy
gcectl set schedule-policy <vm_name> <policy_name> --un
# Example: gcectl set schedule-policy sandbox stop --un

# List all VMs defined in config
gcectl list
```

### Global Flags

- `--config`, `-c` - Config file path (default: "~/.config/gcectl/config.yaml")

### Configuration

Create a `config.yaml` file in your home directory or specify a custom location with the `--config` flag.

#### Configuration Fields

- `default-project`: Your default GCP project ID
- `default-zone`: Default compute zone for operations
- `vm`: List of VM configurations
  - `name`: Name of the VM instance
  - `project`: Project ID where VM resides (overrides default-project)
  - `zone`: Zone where VM resides (overrides default-zone)

#### Example Configuration

```yaml
default-project: your-gcp-project
default-zone: us-central1-a
vm:
  - name: vm-name
    project: your-gcp-project
    zone: us-central1-a
  - name: another-vm
    project: your-gcp-project
    zone: asia-northeast1-a
```

## Directory Structure

```sh
golang/
├── .cobra.yaml         # Cobra CLI framework configuration
├── cmd/                # Command implementations
│   ├── list.go         # List VMs command
│   ├── off.go          # Turn off VM command
│   ├── on.go           # Turn on VM command
│   ├── root.go         # Root command and global flags
│   └── set/            # Set commands
│       ├── machine_type.go  # Set machine type command
│       ├── schedule.go      # Set schedule policy command
│       └── set.go           # Set command group
├── main.go             # Application entry point
├── pkg/                # Core packages
│   ├── config/         # Configuration parsing
│   ├── gce/            # GCE API interaction logic
│   ├── log/            # Logger configuration
│   └── utils/          # Utility functions
└── config.yaml   # Example configuration file
```

### Common Issues

- **Authentication Errors**: Ensure you're authenticated with `gcloud auth login`
- **Permission Denied**: Verify your account has sufficient IAM permissions
- **VM Not Found**: Check VM name and project/zone settings in config

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## TODO

Commands

- [x] on
- [x] off
- [x] set-machine-type
- [x] set-schedule
- [ ] list machine-type
- [ ] list vm

Output Format

- [ ] spin to wait
  - <https://github.com/charmbracelet/bubbletea>

Release

- [ ] Use [GoRelease](https://goreleaser.com/)
