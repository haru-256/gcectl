package gce

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/haru-256/gce-commands/pkg/log"
)

func FetchStatus(ctx context.Context, projectID, zone, instanceName string) (*string, error) {
	// Create a new Instances client
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create Instances client: %v", err)
		return nil, err
	}
	defer instancesClient.Close()

	// Create the request to get instance details
	req := &computepb.GetInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: instanceName,
	}

	// Fetch the instance details
	instance, err := instancesClient.Get(ctx, req)
	if err != nil {
		log.Logger.Errorf("Failed to get instance details: %v", err)
		return nil, err
	}

	return instance.Status, nil
}
