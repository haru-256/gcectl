# CLI Session Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deduplicate CLI command setup by moving config loading, signal-aware context creation, GCP VM repository creation, presenter creation, and cleanup into `go/internal/interface/cli/session.go`.

**Architecture:** Add an `internal/interface/cli` package that represents one CLI command execution session. Commands keep argument validation, usecase selection, and user-facing messages; `cli.Session` owns shared execution resources and lifecycle cleanup. This avoids putting non-command helper files under `go/cmd` while keeping the helper in the existing interface layer.

**Tech Stack:** Go 1.25, Cobra, `go.uber.org/mock` via `go generate`, existing `config`, `gcp`, `log`, `presenter`, `repository`, and `usecase` packages.

---

## File Structure

### Create

- `go/internal/interface/cli/session.go` — session resource bundle and factory functions.
- `go/internal/interface/cli/session_test.go` — unit tests for session creation, dependency injection, and cleanup behavior.
- `go/internal/mock/interface/cli/session_mock.go` — generated mock for `cli.VMRepositoryCloser`; create this only by running `go generate`, not by hand.

### Modify

- `go/cmd/on.go` — replace repeated setup with `cli.NewSession`.
- `go/cmd/off.go` — replace repeated setup with `cli.NewSession`.
- `go/cmd/list.go` — replace repeated setup with `cli.NewSession`; keep current partial-result behavior unchanged.
- `go/cmd/describe.go` — replace repeated setup with `cli.NewSession`.
- `go/cmd/set/machine_type.go` — replace repeated setup with `cli.NewSession`; keep existing config flag lookup.
- `go/cmd/set/schedule.go` — replace repeated setup with `cli.NewSession`; keep existing config flag lookup.

### Do Not Modify

- `go/cmd/root.go`
- `go/cmd/version.go`
- `go/cmd/set/set.go`
- `go/internal/usecase/*`
- `go/internal/infrastructure/gcp/*`
- `go/internal/infrastructure/config/*`

## Naming and Placement Decision

Use:

```text
go/internal/interface/cli/session.go
```

Do not use `go/cmd/bootstrap.go`. The `go/cmd` tree conventionally contains Cobra command definitions; a helper file named `bootstrap.go` there can look like a command and is awkward for `cmd/set` to share without package coupling. `internal/interface/cli/session.go` is available to both `cmd` and `cmd/set`, stays inside the interface layer, and describes the lifecycle accurately: one command execution session with resources that must be closed.

## Public API Shape

```go
session, err := cli.NewSession(cmd, CnfPath)
if err != nil {
	presenter.NewConsolePresenter().Error(err.Error())
	os.Exit(1)
}
defer session.Close()
```

`Session` exposes only resources commands need:

```go
type Session struct {
	Console      *presenter.ConsolePresenter
	Config       *config.Config
	Context      context.Context
	VMRepository repository.VMRepository
}
```

It keeps lifecycle internals private:

```go
stop      context.CancelFunc
closeRepo func() error
```

---

## Task 1: Add `cli.Session`

**Files:**
- Create: `go/internal/interface/cli/session.go`
- Create: `go/internal/interface/cli/session_test.go`
- Generate: `go/internal/mock/interface/cli/session_mock.go`

- [ ] **Step 1: Create the session interface skeleton with `go:generate`**

Create `go/internal/interface/cli/session.go` with only the interface and type skeleton needed for mock generation and failing tests:

```go
package cli

import (
	"context"

	"github.com/haru-256/gcectl/internal/domain/repository"
	"github.com/haru-256/gcectl/internal/infrastructure/config"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/spf13/cobra"
)

//go:generate go tool mockgen -source=$GOFILE -destination=../../mock/interface/cli/session_mock.go -package=mock_cli

type VMRepositoryCloser interface {
	repository.VMRepository
	Close() error
}

type ConfigLoader func(string) (*config.Config, error)

type VMRepositoryFactory func(context.Context, infraLog.Logger) (VMRepositoryCloser, error)

type Options struct {
	LoadConfig      ConfigLoader
	NewVMRepository VMRepositoryFactory
	Logger          infraLog.Logger
}

type Session struct {
	Console      *presenter.ConsolePresenter
	Config       *config.Config
	Context      context.Context
	VMRepository repository.VMRepository
}

func NewSession(cmd *cobra.Command, configPath string) (*Session, error) {
	return nil, nil
}

func NewSessionWithOptions(cmd *cobra.Command, configPath string, opts Options) (*Session, error) {
	return nil, nil
}

func (s *Session) Close() {}
```

- [ ] **Step 2: Generate the mock with `go generate`**

Create the generated mock parent directory first because the repository currently has `go/internal/mock/repository/` but not `go/internal/mock/interface/cli/`:

```bash
mkdir -p go/internal/mock/interface/cli
```

Expected: command exits with status `0`.

Run:

```bash
cd go && go generate ./internal/interface/cli
```

Expected:

```text
```

The command should create:

```text
go/internal/mock/interface/cli/session_mock.go
```

Do not hand-edit the generated file.

- [ ] **Step 3: Write the failing tests using the generated mock**

Create `go/internal/interface/cli/session_test.go`:

```go
package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/haru-256/gcectl/internal/infrastructure/config"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	mockCli "github.com/haru-256/gcectl/internal/mock/interface/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewSessionWithOptionsCreatesSessionAndClosesRepository(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockCli.NewMockVMRepositoryCloser(ctrl)
	repo.EXPECT().Close().Return(nil)

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	session, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			require.Equal(t, "config.yaml", path)
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			require.NotNil(t, ctx)
			require.NotNil(t, logger)
			return repo, nil
		},
		Logger: infraLog.DefaultLogger,
	})

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotNil(t, session.Console)
	require.NotNil(t, session.Config)
	require.NotNil(t, session.Context)
	require.Same(t, repo, session.VMRepository)

	session.Close()
}

func TestNewSessionWithOptionsReturnsConfigError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("config failed")

	session, err := NewSessionWithOptions(&cobra.Command{}, "bad.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			return nil, expectedErr
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			t.Fatal("repository factory should not be called when config loading fails")
			return nil, nil
		},
		Logger: infraLog.DefaultLogger,
	})

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, session)
}

func TestNewSessionWithOptionsReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("repository failed")
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	session, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			require.NotNil(t, ctx)
			return nil, expectedErr
		},
		Logger: infraLog.DefaultLogger,
	})

	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "Failed to create VM repository: repository failed")
	require.Nil(t, session)
}

func TestSessionCloseIsNilSafe(t *testing.T) {
	t.Parallel()

	var session *Session
	require.NotPanics(t, func() {
		session.Close()
	})
}
```

- [ ] **Step 4: Run the new tests to verify they fail against the skeleton**

Run:

```bash
cd go && go test ./internal/interface/cli
```

Expected:

```text
FAIL
expected non-nil session
```

- [ ] **Step 5: Implement `Session`**

Replace `go/internal/interface/cli/session.go` with:

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/haru-256/gcectl/internal/domain/repository"
	"github.com/haru-256/gcectl/internal/infrastructure/config"
	"github.com/haru-256/gcectl/internal/infrastructure/gcp"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/spf13/cobra"
)

//go:generate go tool mockgen -source=$GOFILE -destination=../../mock/interface/cli/session_mock.go -package=mock_cli

type VMRepositoryCloser interface {
	repository.VMRepository
	Close() error
}

type ConfigLoader func(string) (*config.Config, error)

type VMRepositoryFactory func(context.Context, infraLog.Logger) (VMRepositoryCloser, error)

type Options struct {
	LoadConfig      ConfigLoader
	NewVMRepository VMRepositoryFactory
	Logger          infraLog.Logger
}

type Session struct {
	Console      *presenter.ConsolePresenter
	Config       *config.Config
	Context      context.Context
	VMRepository repository.VMRepository

	stop      context.CancelFunc
	closeRepo func() error
}

func NewSession(cmd *cobra.Command, configPath string) (*Session, error) {
	return NewSessionWithOptions(cmd, configPath, Options{
		LoadConfig: config.NewConfig,
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			return gcp.NewVMRepository(ctx, logger)
		},
		Logger: infraLog.DefaultLogger,
	})
}

func NewSessionWithOptions(cmd *cobra.Command, configPath string, opts Options) (*Session, error) {
	if opts.LoadConfig == nil {
		opts.LoadConfig = config.NewConfig
	}
	if opts.NewVMRepository == nil {
		opts.NewVMRepository = func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			return gcp.NewVMRepository(ctx, logger)
		}
	}
	if opts.Logger == nil {
		opts.Logger = infraLog.DefaultLogger
	}

	cfg, err := opts.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)

	vmRepo, err := opts.NewVMRepository(ctx, opts.Logger)
	if err != nil {
		stop()
		return nil, fmt.Errorf("Failed to create VM repository: %w", err)
	}

	return &Session{
		Console:      presenter.NewConsolePresenter(),
		Config:       cfg,
		Context:      ctx,
		VMRepository: vmRepo,
		stop:         stop,
		closeRepo:    vmRepo.Close,
	}, nil
}

func (s *Session) Close() {
	if s == nil {
		return
	}
	if s.closeRepo != nil {
		_ = s.closeRepo()
	}
	if s.stop != nil {
		s.stop()
	}
}
```

- [ ] **Step 6: Regenerate mocks and run the package test**

Run mock generation first so the generated mock stays in sync with `VMRepositoryCloser`:

```bash
cd go && go generate ./internal/interface/cli
```

Expected:

```text
```

Then run:

Run:

```bash
cd go && go test ./internal/interface/cli ./internal/mock/interface/cli
```

Expected:

```text
ok  	github.com/haru-256/gcectl/internal/interface/cli
ok  	github.com/haru-256/gcectl/internal/mock/interface/cli
```

- [ ] **Step 7: Commit this task if commits are requested**

Do not commit unless the user explicitly requested commits. If commits are requested, run:

```bash
git add go/internal/interface/cli/session.go go/internal/interface/cli/session_test.go go/internal/mock/interface/cli/session_mock.go
git commit -m "refactor(cli): add command session setup helper"
```

---

## Task 2: Refactor top-level commands to use `cli.Session`

**Files:**
- Modify: `go/cmd/on.go`
- Modify: `go/cmd/off.go`
- Modify: `go/cmd/describe.go`
- Modify: `go/cmd/list.go`

- [ ] **Step 1: Refactor `go/cmd/on.go`**

In `go/cmd/on.go`, imports should include:

```go
import (
	"context"
	"fmt"
	"os"
	"strings"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)
```

Replace `onRun` with:

```go
func onRun(cmd *cobra.Command, args []string) {
	vmNames := args
	infraLog.DefaultLogger.Debugf("Turning on the instances %s", strings.Join(vmNames, ", "))

	session, err := cli.NewSession(cmd, CnfPath)
	if err != nil {
		presenter.NewConsolePresenter().Error(err.Error())
		os.Exit(1)
	}
	defer session.Close()

	vms, err := session.Config.ResolveVMs(vmNames)
	if err != nil {
		session.Console.Error(err.Error())
		os.Exit(1)
	}

	startVMUseCase := usecase.NewStartVMUseCase(session.VMRepository, infraLog.DefaultLogger)

	err = session.Console.ExecuteWithProgress(
		session.Context,
		fmt.Sprintf("Starting VMs %s", strings.Join(vmNames, ", ")),
		func(ctx context.Context) error {
			return startVMUseCase.Execute(ctx, vms)
		},
	)
	if err != nil {
		session.Console.Error(fmt.Sprintf("Failed to turn on the instances: %v", err))
		os.Exit(1)
	}

	session.Console.Success(fmt.Sprintf("Turned on the instances: %v", strings.Join(vmNames, ", ")))
}
```

- [ ] **Step 2: Refactor `go/cmd/off.go`**

In `go/cmd/off.go`, imports should include:

```go
import (
	"context"
	"fmt"
	"os"
	"strings"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)
```

Replace the setup section inside `offRun` with this pattern while keeping existing messages and VM resolution semantics:

```go
session, err := cli.NewSession(cmd, CnfPath)
if err != nil {
	presenter.NewConsolePresenter().Error(err.Error())
	os.Exit(1)
}
defer session.Close()

vms, err := session.Config.ResolveVMs(vmNames)
if err != nil {
	session.Console.Error(err.Error())
	os.Exit(1)
}

stopVMUseCase := usecase.NewStopVMUseCase(session.VMRepository, infraLog.DefaultLogger)

err = session.Console.ExecuteWithProgress(
	session.Context,
	fmt.Sprintf("Stopping VMs %s", strings.Join(vmNames, ", ")),
	func(ctx context.Context) error {
		return stopVMUseCase.Execute(ctx, vms)
	},
)
```

All later error/success output in `offRun` should use `session.Console`.

- [ ] **Step 3: Refactor `go/cmd/describe.go`**

In `go/cmd/describe.go`, imports should include:

```go
import (
	"fmt"
	"os"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)
```

Inside `Run`, replace config/context/repository setup with:

```go
session, err := cli.NewSession(cmd, CnfPath)
if err != nil {
	presenter.NewConsolePresenter().Error(err.Error())
	os.Exit(1)
}
defer session.Close()

vm, err := session.Config.ResolveVM(vmName)
if err != nil {
	session.Console.Error(err.Error())
	os.Exit(1)
}

describeVMUseCase := usecase.NewDescribeVMUseCase(session.VMRepository)

vmDetail, uptimeStr, err := describeVMUseCase.Execute(session.Context, vm.Project, vm.Zone, vm.Name)
if err != nil {
	session.Console.Error(fmt.Sprintf("Failed to get VM info: %v", err))
	os.Exit(1)
}

session.Console.RenderVMDetail(presenter.VMDetail{
	Name:           vmDetail.Name,
	Project:        vmDetail.Project,
	Zone:           vmDetail.Zone,
	MachineType:    vmDetail.MachineType,
	Status:         vmDetail.Status,
	SchedulePolicy: vmDetail.SchedulePolicy,
	Uptime:         uptimeStr,
})
```

Keep the existing `vmName == ""` validation before creating the session.

- [ ] **Step 4: Refactor `go/cmd/list.go`**

In `go/cmd/list.go`, imports should include:

```go
import (
	"fmt"
	"os"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)
```

Inside `Run`, replace setup with:

```go
session, err := cli.NewSession(cmd, CnfPath)
if err != nil {
	presenter.NewConsolePresenter().Error(err.Error())
	os.Exit(1)
}
defer session.Close()

listVMsUC := usecase.NewListVMsUseCase(session.VMRepository)

items, err := listVMsUC.Execute(session.Context, session.Config.VMs)
infraLog.DefaultLogger.Debugf("Found %d VMs", len(items))
```

Convert and render with `session.Console`:

```go
presenterItems := make([]presenter.VMListItem, len(items))
for i, item := range items {
	presenterItems[i] = presenter.VMListItem{
		Name:           item.VM.Name,
		Project:        item.VM.Project,
		Zone:           item.VM.Zone,
		MachineType:    item.VM.MachineType,
		Status:         item.VM.Status,
		SchedulePolicy: item.VM.SchedulePolicy,
		Uptime:         item.Uptime,
	}
}

if len(presenterItems) > 0 {
	session.Console.RenderVMList(presenterItems)
}
if err != nil {
	session.Console.Error(fmt.Sprintf("Failed to list some VMs: %v", err))
	os.Exit(1)
}
```

Do not change the current partial-result behavior in this refactor.

- [ ] **Step 5: Run tests for top-level command refactor**

Run:

```bash
cd go && go test -tags=ci ./...
```

Expected:

```text
PASS
```

- [ ] **Step 6: Commit this task if commits are requested**

Do not commit unless the user explicitly requested commits. If commits are requested, run:

```bash
git add go/cmd/on.go go/cmd/off.go go/cmd/describe.go go/cmd/list.go
git commit -m "refactor(cmd): use CLI session in top-level commands"
```

---

## Task 3: Refactor `cmd/set` commands to use `cli.Session`

**Files:**
- Modify: `go/cmd/set/machine_type.go`
- Modify: `go/cmd/set/schedule.go`

- [ ] **Step 1: Refactor `go/cmd/set/machine_type.go` imports**

Imports should include:

```go
import (
	"context"
	"fmt"
	"os"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)
```

- [ ] **Step 2: Refactor `go/cmd/set/machine_type.go` body**

Keep the existing config flag lookup because `cmd/set` is a different package and this plan intentionally avoids changing flag ownership:

```go
cnfPath, err := cmd.Flags().GetString("config")
if err != nil {
	console.Error("config is required")
	os.Exit(1)
}
```

Replace config/context/repository setup with:

```go
session, err := cli.NewSession(cmd, cnfPath)
if err != nil {
	presenter.NewConsolePresenter().Error(err.Error())
	os.Exit(1)
}
defer session.Close()

vm, err := session.Config.ResolveVM(vmName)
if err != nil {
	session.Console.Error(err.Error())
	os.Exit(1)
}

updateMachineTypeUseCase := usecase.NewUpdateMachineTypeUseCase(session.VMRepository, infraLog.DefaultLogger)

message := fmt.Sprintf("Updating machine type for VM %s", vmName)
err = session.Console.ExecuteWithProgress(session.Context, message, func(ctx context.Context) error {
	return updateMachineTypeUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, machineType)
})
if err != nil {
	session.Console.Error(fmt.Sprintf("Failed to set machine-type: %v", err))
	os.Exit(1)
}
session.Console.Success(fmt.Sprintf("Set machine-type to %v", machineType))
```

- [ ] **Step 3: Refactor `go/cmd/set/schedule.go` imports**

Imports should include:

```go
import (
	"context"
	"fmt"
	"os"

	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/cli"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/haru-256/gcectl/internal/usecase"
	"github.com/spf13/cobra"
)
```

- [ ] **Step 4: Refactor `go/cmd/set/schedule.go` body**

Keep the existing config flag lookup:

```go
cnfPath, err := cmd.Flags().GetString("config")
if err != nil {
	console.Error("config is required")
	os.Exit(1)
}
```

Replace config/context/repository setup with:

```go
session, err := cli.NewSession(cmd, cnfPath)
if err != nil {
	presenter.NewConsolePresenter().Error(err.Error())
	os.Exit(1)
}
defer session.Close()

vm, err := session.Config.ResolveVM(vmName)
if err != nil {
	session.Console.Error(err.Error())
	os.Exit(1)
}
```

In the `unset` branch, use:

```go
unsetSchedulePolicyUseCase := usecase.NewUnsetSchedulePolicyUseCase(session.VMRepository, infraLog.DefaultLogger)

var message string
if vm.SchedulePolicy != "" {
	message = fmt.Sprintf("Unsetting schedule policy %s for VM %s", vm.SchedulePolicy, vmName)
} else {
	message = fmt.Sprintf("Unsetting schedule policy for VM %s", vmName)
}

err = session.Console.ExecuteWithProgress(session.Context, message, func(ctx context.Context) error {
	return unsetSchedulePolicyUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, policyName)
})
if err != nil {
	session.Console.Error(fmt.Sprintf("Failed to unset schedule-policy: %v", err))
	os.Exit(1)
}
session.Console.Success(fmt.Sprintf("Unset schedule-policy: %v", policyName))
```

In the set branch, use:

```go
setSchedulePolicyUseCase := usecase.NewSetSchedulePolicyUseCase(session.VMRepository, infraLog.DefaultLogger)

message := fmt.Sprintf("Setting schedule policy %s for VM %s", policyName, vmName)

err = session.Console.ExecuteWithProgress(session.Context, message, func(ctx context.Context) error {
	return setSchedulePolicyUseCase.Execute(ctx, vm.Project, vm.Zone, vm.Name, policyName)
})
if err != nil {
	session.Console.Error(fmt.Sprintf("Failed to set schedule-policy: %v", err))
	os.Exit(1)
}
session.Console.Success(fmt.Sprintf("Set schedule-policy: %v", policyName))
```

- [ ] **Step 5: Run tests for set command refactor**

Run:

```bash
cd go && go test -tags=ci ./...
```

Expected:

```text
PASS
```

- [ ] **Step 6: Commit this task if commits are requested**

Do not commit unless the user explicitly requested commits. If commits are requested, run:

```bash
git add go/cmd/set/machine_type.go go/cmd/set/schedule.go
git commit -m "refactor(cmd): use CLI session in set commands"
```

---

## Task 4: Format, test, and verify

**Files:**
- Modify only files touched in Tasks 1-3.

- [ ] **Step 1: Regenerate mocks**

Run:

```bash
mkdir -p go/internal/mock/interface/cli && cd go && go generate ./internal/interface/cli
```

Expected: command exits with status `0` and updates `internal/mock/interface/cli/session_mock.go` only if the interface changed.

- [ ] **Step 2: Format touched non-generated Go files**

Run:

```bash
cd go && gofmt -w \
  internal/interface/cli/session.go \
  internal/interface/cli/session_test.go \
  cmd/on.go \
  cmd/off.go \
  cmd/list.go \
  cmd/describe.go \
  cmd/set/machine_type.go \
  cmd/set/schedule.go
```

Expected: command exits with status `0`.

- [ ] **Step 3: Run all CI-tagged tests**

Run:

```bash
cd go && go test -tags=ci ./...
```

Expected:

```text
PASS
```

- [ ] **Step 4: Run vet**

Run:

```bash
cd go && go vet ./...
```

Expected: no output and exit status `0`.

- [ ] **Step 5: Run lint when available**

Run:

```bash
cd go && golangci-lint run --config=./.golangci.yml ./...
```

Expected: no issues.

If `golangci-lint` is not installed, record the exact command failure and rely on `go test -tags=ci ./...` plus `go vet ./...`.

- [ ] **Step 6: Run command behavior smoke checks**

These checks guard against command entrypoint behavior drift. Run config failure checks from `go/`:

```bash
go run . list --config /tmp/gcectl-missing-config.yaml
go run . describe sandbox --config /tmp/gcectl-missing-config.yaml
go run . on sandbox --config /tmp/gcectl-missing-config.yaml
go run . off sandbox --config /tmp/gcectl-missing-config.yaml
go run . set machine-type sandbox n1-standard-1 --config /tmp/gcectl-missing-config.yaml
go run . set schedule-policy sandbox stop --config /tmp/gcectl-missing-config.yaml
```

Expected for each command:

```text
exit status 1
```

Expected output should contain the config loading error from `config.NewConfig`, not `Failed to create VM repository:`. This verifies that `cli.NewSession` still fails during config loading before creating a repository.

If GCP credentials and a valid config are available, also run one non-mutating happy path:

```bash
go run . list --config /path/to/valid-gcectl-config.yaml
```

Expected: command renders the same VM list format as before the refactor. If one or more VMs fail to load, `list` may still render partial results and then exit with `Failed to list some VMs: ...`; do not change that behavior in this refactor.

- [ ] **Step 7: Commit final verification if commits are requested and there are remaining changes**

Do not commit unless the user explicitly requested commits. If commits are requested and earlier task commits were not made, run:

```bash
git add go/internal/interface/cli/session.go go/internal/interface/cli/session_test.go \
  go/internal/mock/interface/cli/session_mock.go \
  go/cmd/on.go go/cmd/off.go go/cmd/list.go go/cmd/describe.go \
  go/cmd/set/machine_type.go go/cmd/set/schedule.go
git commit -m "refactor(cmd): centralize CLI command session setup"
```

---

## Explicit Non-Goals

- Do not fix `cmd/list.go` partial-error rendering order in this change.
- Do not unify `cmd/set` config flag access with `cmd.CnfPath` in this change.
- Do not remove `os.Exit(1)` from commands.
- Do not introduce a DI framework.
- Do not change usecase constructors or repository interfaces.
- Do not refactor GCP schedule policy duplication.
- Do not put helper code under `go/cmd/bootstrap.go` or any other command-looking file in `go/cmd`.
- Do not hand-write fake implementations of `VMRepositoryCloser`; use the generated `go.uber.org/mock` mock from `go generate`.

## Risks and Unknowns

- Automated command-level tests are not added in this plan because existing command handlers call `os.Exit(1)` directly, which makes isolated command tests intrusive. The smoke checks in Task 4 are required to compensate.
- `Session.Close()` intentionally ignores repository close errors to preserve the existing deferred close behavior in commands.
- `NewSession` must preserve current failure message shape for repository initialization by returning an error whose string begins with `Failed to create VM repository:`. Config loading errors must remain raw `config.NewConfig` errors.
- `Session` exposes several fields. Implementers should use only the field needed by each command and must not move command-specific validation or usecase-specific logic into `internal/interface/cli`.

## Self-Review Notes

- Spec coverage: The plan covers placement, naming, session lifecycle, command migration, tests, and verification.
- Placeholder scan: No `TBD`, `TODO`, or unspecified implementation steps remain.
- Type consistency: The plan consistently uses `Session`, `Options`, `VMRepositoryCloser`, `NewSession`, and `NewSessionWithOptions`.

## Implementation Log
<!-- Implementer appends one line per attempt: [YYYY-MM-DD] attempt #N → STATUS | commit-or-failure-signature -->
[2026-05-09] attempt #1 → DONE | Task 1 complete: session.go, session_test.go, and session_mock.go generated and passing
[2026-05-09] attempt #2 → DONE | Task 1 review fixes: nil cmd.Context() guard, idempotent Close(), added lifecycle tests
[2026-05-09] attempt #3 → DONE | Task 2 complete: refactored on.go, off.go, describe.go, and list.go to use cli.NewSession; go test -tags=ci ./... passes
[2026-05-09] attempt #4 → DONE | Task 3 complete: refactored machine_type.go and schedule.go to use cli.NewSession; go test -tags=ci ./... passes
[2026-05-09] attempt #5 → DONE | Behavior-preservation fix: made VM repository lazy via Session.OpenVMRepository(); updated all commands and tests; go test -tags=ci ./... passes
[2026-05-09] attempt #6 → DONE | Task 4 verification: go generate, gofmt, tests, vet, and lint pass; smoke checks confirm config-error-first behavior. Fixed govet shadow errors (6 files) and added //nolint:staticcheck for capitalized error string preservation.
[2026-05-09] attempt #7 → DONE | Fixed schedule.go to use new Session API (3-return NewSession, OpenVMRepository(ctx), local ctx instead of session.Context); all tests, vet, lint pass.

## Review Findings
<!-- Reviewer appends one line per review: [YYYY-MM-DD] ARTIFACT_TYPE → VERDICT | key issue -->
[2026-05-09] plan → REQUEST_CHANGES | missing command-level behavior-preservation verification; possible repo-init error text drift
[2026-05-09] plan → CHANGES_REQUESTED | prior findings addressed, but config-error test still expects repo-init wrapper text contrary to required raw config error behavior
[2026-05-09] plan → PASS_WITH_NOTES | prior assertion contradiction resolved; no remaining blocker-level test contradictions or compile-obvious snippet issues
[2026-05-09] plan → CHANGES_REQUESTED | go:generate writes to internal/mock/interface/cli/session_mock.go but the plan never creates internal/mock/interface/cli/, so Task 1 likely fails before TDD starts
[2026-05-09] plan → APPROVE | mock directory creation now covers go:generate path, and no remaining blocker-level mock-path, import-cycle, or compile-obvious issues found
[2026-05-09] code-quality-review → FIXED | nil cmd.Context() panics signal.NotifyContext; Close() not idempotent; missing lifecycle tests for nil context/default fallback
[2026-05-09] code-quality-review → FIXED | off.go error message drift: restored original "Failed to turn off the instance(s): %v"
[2026-05-09] code-quality-review → FIXED | os.Exit(1) on error paths after session creation skipped deferred session.Close(); added explicit session.Close() before each os.Exit(1) in on.go, off.go, describe.go, list.go

## Deviations from Plan
<!-- Implementer documents intentional deviations and reasons -->
[2026-05-09] Changed `NewSession` to NOT create VM repository eagerly. Added `Session.OpenVMRepository()` so commands resolve config VMs before touching GCP, preserving original failure ordering and preventing GCP auth/network errors from masking local config/VM errors. This deviates from the original plan's `NewSession` implementation but was required by quality review.

## Open Questions
<!-- Any agent adds questions for orchestrator or arbiter -->
