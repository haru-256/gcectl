# gcectl: Google Cloud Compute Engine Commands

[![Go](https://github.com/haru-256/gcectl/actions/workflows/go.yml/badge.svg)](https://github.com/haru-256/gcectl/actions/workflows/go.yml)

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

### Quick Start

```sh
# List all VMs defined in config (with status and uptime)
gcectl list

# Describe a specific VM (detailed information)
gcectl describe <vm_name>

# Start a VM
gcectl on <vm_name>

# Stop a VM
gcectl off <vm_name>
```

### Advanced Commands

```sh
# Set machine type (VM must be stopped)
gcectl set machine-type <vm_name> <machine-type>
# Example: gcectl set machine-type sandbox e2-medium

# Set schedule policy (auto start/stop)
gcectl set schedule <vm_name> <policy_name>
# Example: gcectl set schedule sandbox my-schedule-policy

# Unset schedule policy
gcectl set schedule <vm_name> <policy_name> --un
# Example: gcectl set schedule sandbox my-schedule-policy --un
```

### Example Output

**List Command:**

```
┌──────────┬────────────┬──────────────┬──────────────┬─────────────┬──────────┬─────────┐
│   Name   │  Project   │     Zone     │ Machine-Type │   Status    │ Schedule │ Uptime  │
├──────────┼────────────┼──────────────┼──────────────┼─────────────┼──────────┼─────────┤
│ my-vm    │ my-project │ us-central1-a│ e2-medium    │ 🟢 RUNNING  │ policy-1 │ 2h30m   │
│ dev-vm   │ my-project │ us-west1-a   │ n1-standard-1│ � RUNNING  │          │ 7d12h45m│
│ test-vm  │ my-project │ asia-east1-a │ e2-small     │ 🟢 RUNNING  │          │ 5m30s   │
│ old-vm   │ my-project │ us-east1-b   │ e2-micro     │ �🔴 STOPPED  │          │ N/A     │
└──────────┴────────────┴──────────────┴──────────────┴─────────────┴──────────┴─────────┘
```

**Describe Command:**

```
• Name          : my-vm
• Project       : my-project
• Zone          : us-central1-a
• MachineType   : e2-medium
• Status        : 🟢 RUNNING
• Uptime        : 2h30m
• SchedulePolicy: my-schedule-policy
```

**Operation with Progress:**

```
$ gcectl on my-vm
Starting VM my-vm...
[SUCCESS] | VM my-vm started successfully
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

## Architecture & Design Philosophy

This project follows **Clean Architecture** principles with clear separation of concerns across multiple layers:

### Design Principles

#### 1. **Layered Architecture**

The application is organized into distinct layers, each with specific responsibilities:

```
┌─────────────────────────────────────────────────────────┐
│                     Interface Layer                      │
│              (cmd/, presenter/)                          │
│  • CLI commands (Cobra)                                  │
│  • Console presentation (lipgloss)                       │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                    Use Case Layer                        │
│                  (usecase/)                              │
│  • Business logic orchestration                          │
│  • Business rule validation (CanStart/CanStop)           │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                    Domain Layer                          │
│                (domain/model/, domain/repository/)       │
│  • Core business entities (VM, Status)                   │
│  • Business rules (CanStart, CanStop, Uptime)            │
│  • Repository interfaces                                 │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                Infrastructure Layer                      │
│           (infrastructure/gcp/, infrastructure/config/)  │
│  • GCP Compute Engine API integration                    │
│  • Configuration file parsing                            │
│  • Logging implementation                                │
└─────────────────────────────────────────────────────────┘
```

#### 2. **Dependency Rule**

- **Inner layers** (Domain, Use Case) have no dependencies on outer layers
- **Outer layers** (Infrastructure, Interface) depend on inner layers
- Dependencies point **inward** only
- Domain layer is completely independent and contains pure business logic

#### 3. **Key Design Decisions**

**Use Case Layer - Applied Judiciously**

- **Used for**: Operations with business logic (Start, Stop, UpdateMachineType)
  - Example: `CanStart()` checks prevent starting an already running VM
  - Example: `CanStop()` checks prevent stopping an already stopped VM
  - Example: Machine type changes require VM to be in STOPPED state
- **Not used for**: Simple read operations (List, Describe)
  - These operations directly use repository without business logic
  - Follows YAGNI (You Aren't Gonna Need It) principle
- **Shared utilities**: Common functions like `calculateUptimeString()` are extracted
  - Used by multiple commands to maintain consistency
  - Reduces code duplication while maintaining clean architecture

**Repository Pattern**

- Abstracts GCP API interactions behind clean interfaces
- Enables parallel execution using `errgroup` for performance
- Provides progress feedback through callback pattern
- Progress callbacks decouple infrastructure from presentation layer

**Presenter Pattern**

- Centralizes all console output formatting
- Separates presentation logic from business logic
- Uses `lipgloss` for styled terminal output
- Implements `Progress()` and `ProgressDone()` for operation feedback

#### 4. **Testing Strategy**

- **Domain Layer**: Unit tests for business rules (CanStart, CanStop, Uptime)
- **Use Case Layer**: Tests with mock repositories for business logic validation
  - Tests for shared utilities like `calculateUptimeString()`
  - Tests for uptime formatting (`formatUptime()`)
    - Supports days, hours, minutes, and seconds
    - Format: `7d12h45m` (days), `2h30m` (hours), `5m30s` (minutes), `45s` (seconds)
  - Tests for describe and list operations
- **Infrastructure Layer**: Integration tests for configuration parsing
- **Presenter Layer**: Output validation tests
  - Progress indicator tests (`Progress()`, `ProgressDone()`, `ProgressStart()`)
  - Status emoji rendering tests
- All tests use table-driven test pattern for clarity and maintainability
- 80+ test cases with race detection enabled

## Directory Structure

```
go/
├── main.go                          # Application entry point
├── config.yaml                      # Example configuration file
├── cmd/                             # Interface Layer - CLI Commands
│   ├── root.go                      # Root command and global flags
│   ├── describe.go                  # Describe VM command
│   ├── list.go                      # List VMs command
│   ├── on.go                        # Start VM command
│   ├── off.go                       # Stop VM command
│   ├── set/                         # Set command group
│   │   ├── set.go                   # Set command root
│   │   ├── machine_type.go          # Set machine type subcommand
│   │   └── schedule.go              # Set/unset schedule policy subcommand
│   └── internal/
│       └── presenter/               # Presentation logic
│           ├── console.go           # Console output formatting
│           └── console_test.go      # Presentation tests
├── internal/
│   ├── domain/                      # Domain Layer - Core Business Logic
│   │   ├── model/                   # Domain entities
│   │   │   ├── vm.go                # VM entity with business rules
│   │   │   └── vm_test.go           # Domain logic tests
│   │   └── repository/              # Repository interfaces
│   │       └── vm_repository.go     # VM repository contract
│   ├── usecase/                     # Use Case Layer - Application Logic
│   │   ├── describe_vm.go           # Describe VM use case
│   │   ├── describe_vm_test.go      # Describe VM tests
│   │   ├── list_vms.go              # List VMs use case
│   │   ├── list_vms_test.go         # List VMs tests
│   │   ├── vm_uptime.go             # Shared uptime calculation
│   │   ├── vm_uptime_test.go        # Uptime calculation tests
│   │   ├── start_vm.go              # Start VM use case
│   │   ├── start_vm_test.go         # Start VM tests
│   │   ├── stop_vm.go               # Stop VM use case
│   │   ├── stop_vm_test.go          # Stop VM tests
│   │   ├── update_machine_type.go   # Update machine type use case
│   │   ├── update_machine_type_test.go
│   │   ├── set_schedule_policy.go   # Set schedule policy use case
│   │   ├── set_schedule_policy_test.go
│   │   ├── unset_schedule_policy.go # Unset schedule policy use case
│   │   └── unset_schedule_policy_test.go
│   ├── infrastructure/              # Infrastructure Layer - External Concerns
│   │   ├── gcp/                     # GCP API integration
│   │   │   └── vm_repository_impl.go # VM repository implementation
│   │   ├── config/                  # Configuration management
│   │   │   ├── config.go            # Config parser
│   │   │   └── config_test.go       # Config parsing tests
│   │   └── log/                     # Logging abstraction
│   │       └── logger.go            # Logger interface and implementation
│   └── interface/                   # Interface Layer - External Interface
│       └── presenter/               # Output presentation
│           ├── console.go           # Console presenter
│           └── console_test.go      # Presenter tests
├── go.mod                           # Go module definition
├── go.sum                           # Go module checksums
├── Makefile                         # Build and development tasks
└── README.md                        # This file
```

### Layer Responsibilities

#### Domain Layer (`internal/domain/`)

- **Purpose**: Core business logic and rules
- **Contains**:
  - VM entity with status management
  - Business rules: `CanStart()`, `CanStop()`, `Uptime()`
  - Repository interfaces (contracts)
- **Dependencies**: None (pure domain logic)
- **Tests**: Unit tests for all business rules

#### Use Case Layer (`internal/usecase/`)

- **Purpose**: Orchestrate business operations
- **Contains**:
  - Application-specific business logic
  - Coordination between domain and infrastructure
- **When to use**:
  - Operations requiring business rule validation
  - Multi-step operations with domain logic
- **When NOT to use**:
  - Simple CRUD operations without business logic
  - Direct data retrieval without validation
- **Dependencies**: Domain layer only
- **Tests**: Mock-based tests with table-driven patterns

#### Infrastructure Layer (`internal/infrastructure/`)

- **Purpose**: External system integrations
- **Contains**:
  - GCP Compute Engine API client
  - Configuration file parsing (YAML)
  - Logger implementation
- **Dependencies**: Domain layer (implements repository interfaces)
- **Key Features**:
  - Parallel execution with `errgroup`
  - Progress indicators with callback pattern
  - `ProgressCallback` type for clean layer separation
  - `SetProgressCallback()` for injecting presentation logic
  - Error handling and retry logic

#### Interface Layer (`cmd/`, `internal/interface/`)

- **Purpose**: User interaction and presentation
- **Contains**:
  - CLI commands (Cobra framework)
  - Console output formatting (lipgloss)
  - VM list and detail rendering
  - Progress indicators
- **Dependencies**: All inner layers
- **Key Features**:
  - Styled terminal output (tables, lists)
  - Status emojis (🟢 RUNNING, 🔴 STOPPED, etc.)
  - Uptime display for running VMs
  - Progress feedback (`Progress()`, `ProgressDone()`)

### Common Issues

- **Authentication Errors**: Ensure you're authenticated with `gcloud auth login`
- **Permission Denied**: Verify your account has sufficient IAM permissions
- **VM Not Found**: Check VM name and project/zone settings in config

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

When contributing, please:

- Follow the existing Clean Architecture structure
- Add tests for new functionality (aim for >80% coverage)
- Update documentation and docstrings
- Run `make lint` and `go test ./...` before submitting
- Keep use case layer lean - only add when business logic is needed

## Development

### Build

```bash
make build
# or
go build -o bin/gcectl .
```

### Test

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests verbosely
go test ./... -v
```

### Lint

```bash
make lint
# Uses golangci-lint with project configuration
```

## Project Status

### Completed Features ✅

- ✅ Start VM (`on` command)
- ✅ Stop VM (`off` command)
- ✅ Describe VM (`describe` command)
- ✅ List VMs (`list` command)
- ✅ Set machine type (`set machine-type` command)
- ✅ Set/unset schedule policy (`set schedule` command)
- ✅ Clean Architecture implementation
- ✅ Comprehensive test coverage (68+ test cases)
- ✅ Progress indicators during operations
- ✅ Parallel execution for list operations
- ✅ Styled console output with lipgloss
- ✅ Uptime display for running VMs
- ✅ Callback pattern for progress feedback

### Planned Features 🔜

- [ ] Interactive TUI mode with bubbletea
- [ ] List available machine types
- [ ] VM cost estimation
- [ ] Batch operations (start/stop multiple VMs)
- [ ] Configuration validation command
- [ ] Export VM details to JSON/YAML
- [ ] GoReleaser integration for releases

## TODO

Commands

- [x] on
- [x] off
- [x] describe  
- [x] list
- [x] set-machine-type
- [x] set-schedule
- [ ] list machine-type
- [ ] cost estimation
- [ ] batch operations

Architecture

- [x] Clean Architecture implementation
- [x] Domain layer with business rules
- [x] Use case layer (applied judiciously)
- [x] Repository pattern
- [x] Presenter pattern
- [x] Comprehensive tests

Output Format

- [x] Styled output with lipgloss
- [x] Progress dots during operations
- [x] Table format for VM list
- [x] Detail format for single VM
- [ ] Interactive TUI with bubbletea

Testing

- [x] Domain layer tests
- [x] Use case layer tests (68+ test cases)
- [x] Infrastructure layer tests
- [x] Presenter layer tests (including progress indicators)
- [ ] Integration tests with GCP emulator
- [ ] E2E tests

Release

- [ ] Use [GoReleaser](https://goreleaser.com/)
- [ ] Multi-platform binaries
- [ ] Homebrew formula
- [ ] Docker image
