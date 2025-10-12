# gcectl: Google Cloud Compute Engine Management CLI

[![Go](https://github.com/haru-256/gcectl/actions/workflows/go.yml/badge.svg)](https://github.com/haru-256/gcectl/actions/workflows/go.yml)
[![Rust](https://github.com/haru-256/gcectl/actions/workflows/rust.yml/badge.svg)](https://github.com/haru-256/gcectl/actions/workflows/rust.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A powerful and elegant CLI tool for managing Google Cloud Compute Engine instances with clean architecture design.

## ✨ Features

- 🚀 **VM Operations**: Start, stop, and monitor GCE instances
- 📊 **Status Monitoring**: View VM status with uptime tracking
- ⚙️ **Machine Type Management**: Change VM configurations on the fly
- 📅 **Schedule Policies**: Automate VM start/stop schedules
- 🎨 **Beautiful Output**: Styled terminal output with tables and emojis
- ⚡ **Parallel Execution**: Fast operations with concurrent API calls
- 🏗️ **Clean Architecture**: Well-structured codebase following best practices
- ✅ **Comprehensive Tests**: 70+ test cases with race detection

## 📦 Installation

### Prerequisites

- Go 1.21 or higher
- Google Cloud SDK configured
- Active GCP project with Compute Engine API enabled
- Appropriate IAM permissions for GCE management

### From Source

```bash
git clone https://github.com/haru-256/gcectl.git
cd gcectl/go
make build
# Binary will be available at bin/gcectl
```

## 🚀 Quick Start

### Configuration

Create a configuration file at `~/.config/gcectl/config.yaml`:

```yaml
default-project: your-gcp-project
default-zone: us-central1-a
vm:
  - name: my-vm
    project: your-gcp-project
    zone: us-central1-a
  - name: dev-vm
    project: your-gcp-project
    zone: asia-northeast1-a
```

### Basic Commands

```bash
# List all VMs with status and uptime
gcectl list

# View detailed information about a VM
gcectl describe my-vm

# Start a VM
gcectl on my-vm

# Stop a VM
gcectl off my-vm

# Change machine type (VM must be stopped)
gcectl set machine-type my-vm e2-medium

# Set schedule policy
gcectl set schedule my-vm my-schedule-policy

# Unset schedule policy
gcectl set schedule my-vm my-schedule-policy --un
```

## 📖 Usage Examples

### List VMs

```bash
gcectl list
```

**Output:**

```
┌──────────┬────────────┬──────────────┬──────────────┬─────────────┬──────────┬─────────┐
│   Name   │  Project   │     Zone     │ Machine-Type │   Status    │ Schedule │ Uptime  │
├──────────┼────────────┼──────────────┼──────────────┼─────────────┼──────────┼─────────┤
│ my-vm    │ my-project │ us-central1-a│ e2-medium    │ 🟢 RUNNING  │ policy-1 │ 2h30m   │
│ dev-vm   │ my-project │ us-west1-a   │ n1-standard-1│ 🔴 STOPPED  │          │ N/A     │
└──────────┴────────────┴──────────────┴──────────────┴─────────────┴──────────┴─────────┘
```

### Describe a VM

```bash
gcectl describe my-vm
```

**Output:**

```
• Name          : my-vm
• Project       : my-project
• Zone          : us-central1-a
• MachineType   : e2-medium
• Status        : 🟢 RUNNING
• Uptime        : 2h30m15s
• SchedulePolicy: my-schedule-policy
```

### Start a VM

```bash
$ gcectl on my-vm
Starting VM my-vm...
[SUCCESS] | VM my-vm started successfully
```

### Change Machine Type

```bash
$ gcectl set machine-type my-vm e2-standard-2
Updating machine type for VM my-vm...
[SUCCESS] | Set machine-type to e2-standard-2
```

## 🏗️ Architecture

This project follows **Clean Architecture** principles with strict layer separation:

```
┌─────────────────────────────────────────────────────────┐
│                   Interface Layer                        │
│           (cmd/, presenter/)                             │
│   • CLI Commands (Cobra)                                 │
│   • Console Presentation (lipgloss)                      │
│   • Progress Indicators                                  │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                  Use Case Layer                          │
│                 (usecase/)                               │
│   • Business Logic Orchestration                         │
│   • VM Operations (Start, Stop, Update)                  │
│   • Data Retrieval (List, Describe)                      │
│   • Shared Utilities (Uptime Calculation)                │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                   Domain Layer                           │
│         (domain/model/, domain/repository/)              │
│   • Core Entities (VM, Status)                           │
│   • Business Rules (CanStart, CanStop)                   │
│   • Repository Interfaces                                │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│              Infrastructure Layer                        │
│      (infrastructure/gcp/, infrastructure/config/)       │
│   • GCP Compute Engine API Client                        │
│   • Configuration Management (YAML)                      │
│   • Progress Callback Implementation                     │
│   • Logging & Error Handling                             │
└─────────────────────────────────────────────────────────┘
```

### Key Design Principles

- **Dependency Rule**: Dependencies point inward only
- **Layer Independence**: Inner layers have no knowledge of outer layers
- **Callback Pattern**: Clean separation between infrastructure and presentation
- **Repository Pattern**: Abstract external API interactions
- **YAGNI**: Use cases applied only where business logic exists

For detailed architecture documentation, see [go/README.md](go/README.md).

## 🧪 Testing

The project maintains high test coverage with comprehensive test suites:

```bash
cd go

# Run all tests
make test

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run specific test package
go test ./internal/usecase/... -v
```

**Test Coverage:**

- ✅ 70+ test cases
- ✅ Domain layer: Business rule tests
- ✅ Use case layer: Mock-based integration tests
- ✅ Infrastructure layer: Configuration parsing tests
- ✅ Presenter layer: Output validation tests
- ✅ Race detection enabled
- ✅ Table-driven test patterns

## 🛠️ Development

### Build

```bash
cd go
make build
# Output: bin/gcectl
```

### Lint

```bash
cd go
make lint
# Uses golangci-lint with strict configuration
```

### Project Structure

```
gcectl/
├── go/                              # Go implementation
│   ├── cmd/                         # CLI commands
│   │   ├── describe.go              # Describe VM command
│   │   ├── list.go                  # List VMs command
│   │   ├── on.go                    # Start VM command
│   │   ├── off.go                   # Stop VM command
│   │   ├── root.go                  # Root command
│   │   └── set/                     # Set command group
│   │       ├── machine_type.go      # Set machine type
│   │       └── schedule.go          # Set/unset schedule
│   ├── internal/
│   │   ├── domain/                  # Domain layer
│   │   │   ├── model/               # Entities (VM)
│   │   │   └── repository/          # Repository interfaces
│   │   ├── usecase/                 # Use case layer
│   │   │   ├── describe_vm.go       # Describe VM use case
│   │   │   ├── list_vms.go          # List VMs use case
│   │   │   ├── vm_uptime.go         # Shared uptime logic
│   │   │   ├── start_vm.go          # Start VM use case
│   │   │   ├── stop_vm.go           # Stop VM use case
│   │   │   └── update_machine_type.go
│   │   ├── infrastructure/          # Infrastructure layer
│   │   │   ├── gcp/                 # GCP API client
│   │   │   ├── config/              # Configuration
│   │   │   └── log/                 # Logging
│   │   └── interface/               # Interface layer
│   │       └── presenter/           # Console presenter
│   ├── main.go                      # Application entry
│   ├── config.yaml                  # Example config
│   └── Makefile                     # Build automation
│
├── terraform/                       # Infrastructure as Code
│   ├── environments/dev/            # Dev environment
│   └── modules/                     # Reusable modules
│       ├── gce/                     # GCE instance module
│       └── tfstate_gcs_bucket/      # State bucket module
│
└── rust/                            # Rust implementation (WIP)
```

## 🌟 Status Indicators

The CLI uses emoji indicators for quick status recognition:

- 🟢 **RUNNING** - VM is running
- 🔴 **STOPPED** - VM is stopped
- 🟡 **STAGING** - VM is being staged
- 🟠 **PROVISIONING** - VM is provisioning
- 🔵 **STOPPING** - VM is stopping
- ⚫ **TERMINATED** - VM is terminated
- ⚪ **SUSPENDING** - VM is suspending
- 🟤 **SUSPENDED** - VM is suspended
- 🔄 **REPAIRING** - VM is being repaired

## 📚 Additional Resources

- **Go Implementation**: See [go/README.md](go/README.md) for detailed documentation
- **Terraform**: Infrastructure provisioning configurations in [terraform/](terraform/)
- **Architecture Deep Dive**: [go/README.md#architecture--design-philosophy](go/README.md#architecture--design-philosophy)

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Follow Clean Architecture**: Respect layer boundaries
2. **Add Tests**: Aim for >80% coverage for new code
3. **Update Documentation**: Keep README and docstrings current
4. **Run Quality Checks**: Ensure `make test` and `make lint` pass
5. **Keep Use Cases Lean**: Add use case layer only when business logic exists

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details (if available).

## 📋 Roadmap

### Completed ✅

- [x] Start/Stop VM operations
- [x] List VMs with status and uptime
- [x] Describe VM details
- [x] Set machine type
- [x] Set/unset schedule policies
- [x] Clean Architecture implementation
- [x] Progress indicators
- [x] Parallel execution
- [x] Comprehensive test coverage
- [x] Styled console output

### Planned 🔜

- [ ] Interactive TUI mode (bubbletea)
- [ ] List available machine types
- [ ] VM cost estimation
- [ ] Batch operations (multiple VMs)
- [ ] Configuration validation command
- [ ] Export VM details (JSON/YAML)
- [ ] GoReleaser for multi-platform releases
- [ ] Homebrew formula
- [ ] Docker image
- [ ] Integration tests with GCP emulator

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**haru-256**

- GitHub: [@haru-256](https://github.com/haru-256)

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Google Cloud Go SDK](https://github.com/googleapis/google-cloud-go) - GCP API client

---

**Made with ❤️ and Clean Architecture**
