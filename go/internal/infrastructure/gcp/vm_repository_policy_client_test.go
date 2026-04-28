package gcp

import (
	"context"
	"errors"
	"testing"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/stretchr/testify/require"
)

type fakeInstancesClient struct {
	instance *computepb.Instance
	closed   bool
	closeErr error
}

func (c *fakeInstancesClient) Get(context.Context, *computepb.GetInstanceRequest, ...gax.CallOption) (*computepb.Instance, error) {
	return c.instance, nil
}

func (c *fakeInstancesClient) Start(context.Context, *computepb.StartInstanceRequest, ...gax.CallOption) (*compute.Operation, error) {
	return nil, nil
}

func (c *fakeInstancesClient) Stop(context.Context, *computepb.StopInstanceRequest, ...gax.CallOption) (*compute.Operation, error) {
	return nil, nil
}

func (c *fakeInstancesClient) AddResourcePolicies(context.Context, *computepb.AddResourcePoliciesInstanceRequest, ...gax.CallOption) (*compute.Operation, error) {
	return nil, nil
}

func (c *fakeInstancesClient) RemoveResourcePolicies(context.Context, *computepb.RemoveResourcePoliciesInstanceRequest, ...gax.CallOption) (*compute.Operation, error) {
	return nil, nil
}

func (c *fakeInstancesClient) SetMachineType(context.Context, *computepb.SetMachineTypeInstanceRequest, ...gax.CallOption) (*compute.Operation, error) {
	return nil, nil
}

func (c *fakeInstancesClient) Close() error {
	c.closed = true
	return c.closeErr
}

type fakeResourcePoliciesClient struct {
	policy   *computepb.ResourcePolicy
	closed   bool
	closeErr error
}

func (c *fakeResourcePoliciesClient) Get(context.Context, *computepb.GetResourcePolicyRequest, ...gax.CallOption) (*computepb.ResourcePolicy, error) {
	return c.policy, nil
}

func (c *fakeResourcePoliciesClient) Close() error {
	c.closed = true
	return c.closeErr
}

func TestVMRepositoryCloseClosesInjectedClients(t *testing.T) {
	instancesClient := &fakeInstancesClient{}
	policyClient := &fakeResourcePoliciesClient{}
	repo := newVMRepository(log.NewLogger(), instancesClient, policyClient)

	require.NoError(t, repo.Close())
	require.True(t, instancesClient.closed)
	require.True(t, policyClient.closed)
}

func TestVMRepositoryCloseReturnsJoinedErrorsAndClosesBothClients(t *testing.T) {
	instancesErr := errors.New("instances close failed")
	policyErr := errors.New("policy close failed")
	instancesClient := &fakeInstancesClient{closeErr: instancesErr}
	policyClient := &fakeResourcePoliciesClient{closeErr: policyErr}
	repo := newVMRepository(log.NewLogger(), instancesClient, policyClient)

	err := repo.Close()
	require.ErrorIs(t, err, instancesErr)
	require.ErrorIs(t, err, policyErr)
	require.True(t, instancesClient.closed)
	require.True(t, policyClient.closed)
}

func TestVMRepositoryFindByNameUsesInjectedInstancesClient(t *testing.T) {
	instancesClient := &fakeInstancesClient{
		instance: &computepb.Instance{
			Name:        stringPtr("sandbox-1"),
			SelfLink:    stringPtr("https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a/instances/sandbox-1"),
			Zone:        stringPtr("https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a"),
			Status:      stringPtr("RUNNING"),
			MachineType: stringPtr("https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a/machineTypes/e2-medium"),
		},
	}
	policyClient := &fakeResourcePoliciesClient{}
	repo := newVMRepository(log.NewLogger(), instancesClient, policyClient)

	vm, err := repo.FindByName(context.Background(), &model.VM{
		Project: "test-project",
		Zone:    "us-central1-a",
		Name:    "sandbox-1",
	})
	require.NoError(t, err)
	require.Equal(t, "sandbox-1", vm.Name)
	require.Equal(t, "test-project", vm.Project)
	require.Equal(t, "us-central1-a", vm.Zone)
}
