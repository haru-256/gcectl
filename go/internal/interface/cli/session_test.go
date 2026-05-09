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

func TestNewSessionWithOptionsCreatesSession(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	session, ctx, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			require.Equal(t, "config.yaml", path)
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			t.Fatal("repository factory should not be called during NewSession")
			return nil, nil
		},
		Logger: infraLog.DefaultLogger,
	})

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotNil(t, ctx)
	require.NotNil(t, session.Console)
	require.NotNil(t, session.Config)
	require.Nil(t, session.VMRepository)
}

func TestNewSessionWithOptionsReturnsConfigError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("config failed")

	session, ctx, err := NewSessionWithOptions(&cobra.Command{}, "bad.yaml", Options{
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
	require.Nil(t, ctx)
}

func TestNewSessionWithOptionsHandlesNilCommandContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockCli.NewMockVMRepositoryCloser(ctrl)
	repo.EXPECT().Close().Return(nil)

	cmd := &cobra.Command{}

	session, ctx, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			require.NotNil(t, ctx)
			return repo, nil
		},
		Logger: infraLog.DefaultLogger,
	})

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotNil(t, ctx)

	err = session.OpenVMRepository(ctx)
	require.NoError(t, err)
	session.Close()
}

func TestNewSessionWithOptionsFallsBackToDefaultOptions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockCli.NewMockVMRepositoryCloser(ctrl)
	repo.EXPECT().Close().Return(nil)

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	session, ctx, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			require.NotNil(t, ctx)
			require.NotNil(t, logger)
			return repo, nil
		},
		Logger: nil,
	})

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotNil(t, ctx)

	err = session.OpenVMRepository(ctx)
	require.NoError(t, err)
	session.Close()
}

func TestOpenVMRepositoryCreatesAndStoresRepository(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockCli.NewMockVMRepositoryCloser(ctrl)
	repo.EXPECT().Close().Return(nil)

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	session, ctx, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
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
	require.Nil(t, session.VMRepository)
	require.NotNil(t, ctx)

	err = session.OpenVMRepository(ctx)
	require.NoError(t, err)
	require.Same(t, repo, session.VMRepository)

	session.Close()
}

func TestOpenVMRepositoryReturnsWrappedError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("repository failed")
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	session, ctx, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			require.NotNil(t, ctx)
			return nil, expectedErr
		},
		Logger: infraLog.DefaultLogger,
	})

	require.NoError(t, err)
	require.NotNil(t, ctx)

	err = session.OpenVMRepository(ctx)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "Failed to create VM repository: repository failed")
}

func TestOpenVMRepositoryIsIdempotent(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockCli.NewMockVMRepositoryCloser(ctrl)
	repo.EXPECT().Close().Return(nil)

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	callCount := 0
	session, ctx, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			callCount++
			return repo, nil
		},
		Logger: infraLog.DefaultLogger,
	})

	require.NoError(t, err)
	require.NotNil(t, ctx)

	err = session.OpenVMRepository(ctx)
	require.NoError(t, err)
	require.Same(t, repo, session.VMRepository)
	require.Equal(t, 1, callCount)

	err = session.OpenVMRepository(ctx)
	require.NoError(t, err)
	require.Same(t, repo, session.VMRepository)
	require.Equal(t, 1, callCount)

	session.Close()
}

func TestSessionCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockCli.NewMockVMRepositoryCloser(ctrl)
	repo.EXPECT().Close().Return(nil).Times(1)

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	session, ctx, err := NewSessionWithOptions(cmd, "config.yaml", Options{
		LoadConfig: func(path string) (*config.Config, error) {
			return &config.Config{}, nil
		},
		NewVMRepository: func(ctx context.Context, logger infraLog.Logger) (VMRepositoryCloser, error) {
			return repo, nil
		},
		Logger: infraLog.DefaultLogger,
	})

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotNil(t, ctx)

	err = session.OpenVMRepository(ctx)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		session.Close()
		session.Close()
	})
}

func TestSessionCloseIsNilSafe(t *testing.T) {
	t.Parallel()

	var session *Session
	require.NotPanics(t, func() {
		session.Close()
	})
}
