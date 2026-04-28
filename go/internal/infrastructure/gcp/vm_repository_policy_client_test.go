package gcp

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/stretchr/testify/require"
)

type fakeResourcePoliciesClient struct {
	closed bool
}

func (c *fakeResourcePoliciesClient) Get(context.Context, *computepb.GetResourcePolicyRequest, ...gax.CallOption) (*computepb.ResourcePolicy, error) {
	return nil, nil
}

func (c *fakeResourcePoliciesClient) Close() error {
	c.closed = true
	return nil
}

func TestLazyResourcePoliciesClientProviderRetriesAfterInitializationFailure(t *testing.T) {
	firstErr := errors.New("transient init failure")
	successClient := &fakeResourcePoliciesClient{}
	attempts := 0

	provider := newLazyResourcePoliciesClientProvider(func(context.Context) (resourcePoliciesClient, error) {
		attempts++
		if attempts == 1 {
			return nil, firstErr
		}
		return successClient, nil
	})

	client, err := provider.Get(context.Background())
	require.ErrorIs(t, err, firstErr)
	require.Nil(t, client)

	client, err = provider.Get(context.Background())
	require.NoError(t, err)
	require.Same(t, successClient, client)
	require.Equal(t, 2, attempts)
}

func TestLazyResourcePoliciesClientProviderReusesInitializedClient(t *testing.T) {
	successClient := &fakeResourcePoliciesClient{}
	attempts := 0
	provider := newLazyResourcePoliciesClientProvider(func(context.Context) (resourcePoliciesClient, error) {
		attempts++
		return successClient, nil
	})

	firstClient, err := provider.Get(context.Background())
	require.NoError(t, err)
	secondClient, err := provider.Get(context.Background())
	require.NoError(t, err)

	require.Same(t, successClient, firstClient)
	require.Same(t, successClient, secondClient)
	require.Equal(t, 1, attempts)
}

func TestLazyResourcePoliciesClientProviderCloseClosesInitializedClient(t *testing.T) {
	successClient := &fakeResourcePoliciesClient{}
	provider := newLazyResourcePoliciesClientProvider(func(context.Context) (resourcePoliciesClient, error) {
		return successClient, nil
	})

	client, err := provider.Get(context.Background())
	require.NoError(t, err)
	require.Same(t, successClient, client)

	require.NoError(t, provider.Close())
	require.True(t, successClient.closed)
}

func TestLazyResourcePoliciesClientProviderGetAfterCloseReturnsError(t *testing.T) {
	provider := newLazyResourcePoliciesClientProvider(func(context.Context) (resourcePoliciesClient, error) {
		return &fakeResourcePoliciesClient{}, nil
	})

	require.NoError(t, provider.Close())

	client, err := provider.Get(context.Background())
	require.Error(t, err)
	require.EqualError(t, err, "resource policies client provider is closed")
	require.Nil(t, client)
}

type fakeResourcePoliciesClientProvider struct {
	client resourcePoliciesClient
}

func (p *fakeResourcePoliciesClientProvider) Get(context.Context) (resourcePoliciesClient, error) {
	return p.client, nil
}

func (p *fakeResourcePoliciesClientProvider) Close() error {
	return nil
}

func TestVMRepositoryGetPolicyClientUsesInjectedProvider(t *testing.T) {
	successClient := &fakeResourcePoliciesClient{}
	provider := &fakeResourcePoliciesClientProvider{client: successClient}
	repo := newVMRepository(log.NewLogger(), provider)

	client, err := repo.getPolicyClient(context.Background())
	require.NoError(t, err)
	require.Same(t, successClient, client)
}
