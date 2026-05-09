package cli

import (
	"context"
	"errors"
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
	VMRepository repository.VMRepository

	stop            context.CancelFunc
	closeRepo       func() error
	newVMRepository VMRepositoryFactory
	logger          infraLog.Logger
}

func NewSession(cmd *cobra.Command, configPath string) (*Session, context.Context, error) {
	return NewSessionWithOptions(cmd, configPath, Options{
		LoadConfig: config.NewConfig,
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			return gcp.NewVMRepository(ctx, logger)
		},
		Logger: infraLog.DefaultLogger,
	})
}

func NewSessionWithOptions(cmd *cobra.Command, configPath string, opts Options) (*Session, context.Context, error) {
	if cmd == nil {
		return nil, nil, errors.New("cmd is required")
	}
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
		return nil, nil, err
	}

	parentCtx := cmd.Context()
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, stop := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)

	return &Session{
		Console:         presenter.NewConsolePresenter(),
		Config:          cfg,
		stop:            stop,
		newVMRepository: opts.NewVMRepository,
		logger:          opts.Logger,
	}, ctx, nil
}

func (s *Session) OpenVMRepository(ctx context.Context) error {
	if s == nil {
		return errors.New("session is nil")
	}
	if s.VMRepository != nil || s.closeRepo != nil {
		return nil
	}
	repo, err := s.newVMRepository(ctx, s.logger)
	if err != nil {
		return fmt.Errorf("Failed to create VM repository: %w", err) //nolint:staticcheck // Capitalized error preserved for backward-compatible behavior
	}
	s.VMRepository = repo
	s.closeRepo = repo.Close
	return nil
}

func (s *Session) Close() {
	if s == nil {
		return
	}
	if s.closeRepo != nil {
		_ = s.closeRepo()
		s.closeRepo = nil
	}
	if s.stop != nil {
		s.stop()
		s.stop = nil
	}
}
