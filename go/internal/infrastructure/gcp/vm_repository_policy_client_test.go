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

func TestVMRepository_getPolicyClientRetriesAfterInitializationFailure(t *testing.T) {
	repo := NewVMRepository(log.NewLogger())
	firstErr := errors.New("transient init failure")
	successClient := &fakeResourcePoliciesClient{}
	attempts := 0

	repo.newResourcePoliciesClient = func(context.Context) (resourcePoliciesClient, error) {
		attempts++
		if attempts == 1 {
			return nil, firstErr
		}
		return successClient, nil
	}

	client, err := repo.getPolicyClient(context.Background())
	require.ErrorIs(t, err, firstErr)
	require.Nil(t, client)

	client, err = repo.getPolicyClient(context.Background())
	require.NoError(t, err)
	require.Same(t, successClient, client)
	require.Equal(t, 2, attempts)
}
