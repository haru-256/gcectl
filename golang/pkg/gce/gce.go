package gce

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	farm "github.com/dgryski/go-farm"
	"github.com/haru-256/gcectl/pkg/config"
	"github.com/haru-256/gcectl/pkg/log"
	"golang.org/x/sync/errgroup"
)

// getStatus returns the status of a Compute Engine instance.
//
// Parameters:
//   - ctx: Context for the API call.
//   - instance: A pointer to a computepb.Instance object representing the GCE instance.
//
// Returns:
//   - A string representing the current status of the instance. if the status is empty, it returns "UNKNOWN".
//   - An error, which is always nil in the current implementation.
func getStatus(ctx context.Context, instance *computepb.Instance) (string, error) {
	status := instance.GetStatus()
	if status == "" {
		return "UNKNOWN", fmt.Errorf("status is missing in the instance")
	}
	return status, nil
}

// getMachineType retrieves the machine type of a Google Cloud Compute Engine instance.
// If the machine type is not found, it returns "UNKNOWN" and an error.
func getMachineType(ctx context.Context, instance *computepb.Instance) (string, error) {
	fullURI := instance.GetMachineType()
	if fullURI == "" {
		return "UNKNOWN", fmt.Errorf("machine type is missing in the instance")
	}

	// Extract the machine type name from the full URL
	pattern := `machineTypes/([^/]+)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(fullURI)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract machineType from instance self link")
	}
	machineType := matches[1]
	return machineType, nil
}

// getSchedulePolicy retrieves the schedule policy attached to a GCE instance.
//
// It connects to the Google Cloud Resource Policies API and examines all resource policies
// attached to the given instance. It specifically looks for policies with an InstanceSchedulePolicy
// and returns the name of the first such policy found.
//
// Parameters:
//   - ctx: Context for the API request
//   - instance: Pointer to a computepb.Instance object representing the GCE instance
//
// Returns:
//   - string: The name of the schedule policy (empty string if none is found)
//   - error: Any error encountered during the process
//
// The function iterates through all policies attached to the instance, examines each one,
// and identifies those that contain instance scheduling information.
func getSchedulePolicy(ctx context.Context, instance *computepb.Instance) (string, error) {
	// Create a new ResourcePolicies client
	policyClient, err := compute.NewResourcePoliciesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create Instances client: %v", err)
		return "", err
	}
	defer policyClient.Close()
	policies := instance.GetResourcePolicies()
	project, err := getProjectFromInstance(instance)
	if err != nil {
		log.Logger.Errorf("Failed to get project from instance: %v", err)
		return "", err
	}
	region, err := getRegionFromInstance(instance)
	if err != nil {
		log.Logger.Errorf("Failed to get region from instance: %v", err)
		return "", err
	}

	var schedulePolicyName string = ""
	for _, policy := range policies {
		log.Logger.Debugf("Resource Policy: %s", policy)

		// Extract the policy name from the full URL
		policyParts := strings.Split(policy, "/")
		policyName := policyParts[len(policyParts)-1]

		policyReq := &computepb.GetResourcePolicyRequest{
			Project:        project,
			Region:         region,
			ResourcePolicy: policyName,
		}
		resourcePolicy, er := policyClient.Get(ctx, policyReq)
		if er != nil {
			log.Logger.Errorf("Failed to get resource policy details: %v", er)
			continue
		}
		// Check if the policy has an instance schedule policy
		schedulePolicy := resourcePolicy.GetInstanceSchedulePolicy()
		if schedulePolicy != nil {
			schedulePolicyName = fmt.Sprintf("%s(%s)", policyName, *schedulePolicy.VmStopSchedule.Schedule)
		}
	}
	return schedulePolicyName, nil
}

// getInstance retrieves a specific Google Cloud Compute Engine instance details.
//
// This function creates clients for Instance and ResourcePolicy services, then uses the
// Instance client to fetch details for the specified instance. It returns the instance
// data as a computepb.Instance object.
//
// Parameters:
//   - ctx: The context.Context for the request, used for deadline and cancellation control
//   - projectID: The Google Cloud project ID containing the instance
//   - zone: The Google Cloud zone where the instance is located
//   - instanceName: The name of the instance to retrieve
//
// Returns:
//   - *computepb.Instance: The instance data if successful
//   - error: An error if the operation failed
func getInstance(ctx context.Context, projectID, zone, instanceName string) (*computepb.Instance, error) {
	// Create a new Instances client
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create Instances client: %v", err)
		return nil, err
	}
	defer instancesClient.Close()
	// Create a new ResourcePolicies client
	policyClient, err := compute.NewResourcePoliciesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create Instances client: %v", err)
		return nil, err
	}
	defer policyClient.Close()

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

	return instance, nil
}

// getVMKey generates a unique identifier for a VM based on its name, project, and zone.
// It returns a 64-bit unsigned integer hash using the farm hash algorithm.
// This function is useful for creating consistent keys when caching or tracking VMs.
func getVMKey(name, project, zone string) uint64 {
	return farm.Fingerprint64([]byte(name + project + zone))
}

func UpdateInstancesInfo(ctx context.Context, vms []*config.VM) error {
	// get instances
	instances := map[uint64]*computepb.Instance{}
	for _, vm := range vms {
		instance, err := getInstance(ctx, vm.Project, vm.Zone, vm.Name)
		if err != nil {
			log.Logger.Errorf("Failed to get instance details: %v", err)
			return err // NOTE: We don't want to fail the entire operation for one VM
		}
		key := getVMKey(vm.Name, vm.Project, vm.Zone)
		instances[key] = instance
	}

	// get status and schedule policy
	eg, ctx := errgroup.WithContext(ctx)
	for _, vm := range vms {
		vm := vm
		eg.Go(func() error {
			key := getVMKey(vm.Name, vm.Project, vm.Zone)
			instance, ok := instances[key]
			if !ok {
				log.Logger.Errorf("Instance not found for VM: %s", vm.Name)
				return nil
			}

			status, err := getStatus(ctx, instance)
			vm.Status = status
			if err != nil {
				log.Logger.Errorf("Failed to get status: %v", err)
				// We don't want to fail the entire operation for one VM because errorgroup will stop all goroutines
			}

			machineType, err := getMachineType(ctx, instance)
			vm.MachineType = machineType
			if err != nil {
				log.Logger.Errorf("Failed to get machine-type: %v", err)
			}

			schedulePolicy, err := getSchedulePolicy(ctx, instance)
			if schedulePolicy == "" {
				vm.SchedulePolicy = "#NONE"
			} else {
				vm.SchedulePolicy = schedulePolicy
			}
			if err != nil {
				log.Logger.Errorf("Failed to get schedule policy: %v", err)
			}
			return nil
		})
	}
	// Wait for all goroutines to complete
	if err := eg.Wait(); err != nil {
		log.Logger.Error("Error fetching VM statuses:", err)
		return err
	}

	return nil
}

// getRegionFromInstance extracts the region from a Google Compute Engine instance.
func getRegionFromInstance(instance *computepb.Instance) (string, error) {
	if instance == nil {
		return "", fmt.Errorf("instance is nil")
	}
	// Extract the zone URI from the instance
	// https://www.googleapis.com/compute/v1/projects/{project}/zones/{zone}
	zoneURI := instance.GetZone()
	if zoneURI == "" {
		return "", fmt.Errorf("zone information is missing in the instance")
	}

	// Split the zone URI to get the zone name
	parts := strings.Split(zoneURI, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid zone URI format")
	}
	zoneName := parts[len(parts)-1]

	// Split the zone name to derive the region
	zoneParts := strings.Split(zoneName, "-")
	if len(zoneParts) < 2 {
		return "", fmt.Errorf("invalid zone name format")
	}
	region := strings.Join(zoneParts[:len(zoneParts)-1], "-")

	return region, nil
}

// getProjectFromInstance extracts the project ID from a Google Compute Engine instance.
func getProjectFromInstance(instance *computepb.Instance) (string, error) {
	// Get the self link of the instance
	// https://www.googleapis.com/compute/v1/projects/{project}/zones/{zone}/instances/{instance}
	selfLink := instance.GetSelfLink()
	pattern := `projects/([^/]+)/`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(selfLink)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract project from instance self link")
	}
	project := matches[1]
	return project, nil
}

func OnVM(vm *config.VM) error {
	ctx := context.Background()

	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create instances client: %v", err)
		return err
	}
	defer client.Close()

	// Create the Start request
	req := &computepb.StartInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	// Execute the request
	op, err := client.Start(ctx, req)
	if err != nil {
		log.Logger.Errorf("failed to start instance: %v", err)
		return err
	}

	fmt.Printf("Turn ON Instance %s", vm.Name)

	// wait for the operation to complete
	if err = waitOperator(ctx, op); err != nil {
		log.Logger.Errorf("failed to start instance: %v", err)
		return err
	}

	return nil
}

func OffVM(vm *config.VM) error {
	ctx := context.Background()

	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create instances client: %v", err)
		return err
	}
	defer client.Close()

	// Create the Start request
	req := &computepb.StopInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	// Execute the request
	op, err := client.Stop(ctx, req)
	if err != nil {
		log.Logger.Errorf("failed to start instance: %v", err)
		return err
	}

	fmt.Printf("Turn OFF Instance %s", vm.Name)

	// wait for the operation to complete
	if err = waitOperator(ctx, op); err != nil {
		log.Logger.Errorf("failed to stop instance: %v", err)
		return err
	}

	return nil
}

func SetMachineType(vm *config.VM, machineType string) error {
	ctx := context.Background()
	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create instances client: %v", err)
		return err
	}
	defer client.Close()
	// Set the new machine type
	setMachineTypeReq := &computepb.SetMachineTypeInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
		InstancesSetMachineTypeRequestResource: &computepb.InstancesSetMachineTypeRequest{
			MachineType: toPtr(fmt.Sprintf("zones/%s/machineTypes/%s", vm.Zone, machineType)),
		},
	}
	op, err := client.SetMachineType(ctx, setMachineTypeReq)
	if err != nil {
		log.Logger.Errorf("Failed to set machine type: %v", err)
		return err
	}

	fmt.Printf("Setting machine type to %s for instance %s", machineType, vm.Name)

	// wait for the operation to complete
	if err = waitOperator(ctx, op); err != nil {
		log.Logger.Errorf("Failed to set machine type: %v", err)
		return err
	}

	return nil
}

// SetSchedulePolicy attaches a schedule policy to a Google Compute Engine instance.
func SetSchedulePolicy(vm *config.VM, policyName string) error {
	ctx := context.Background()
	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("failed to create Instances client: %v", err)
		return err
	}
	defer client.Close()

	instance, err := getInstance(ctx, vm.Project, vm.Zone, vm.Name)
	if err != nil {
		log.Logger.Errorf("failed to get instance: %v", err)
		return err
	}
	region, err := getRegionFromInstance(instance)
	if err != nil {
		log.Logger.Errorf("Failed to get region from instance: %v", err)
		return err
	}

	policySelfLink := fmt.Sprintf("projects/%s/regions/%s/resourcePolicies/%s", vm.Project, region, policyName)

	req := &computepb.AddResourcePoliciesInstanceRequest{
		Instance: vm.Name,
		Project:  vm.Project,
		Zone:     vm.Zone,
		InstancesAddResourcePoliciesRequestResource: &computepb.InstancesAddResourcePoliciesRequest{
			ResourcePolicies: []string{policySelfLink},
		},
	}

	op, err := client.AddResourcePolicies(ctx, req)
	if err != nil {
		log.Logger.Errorf("Failed to set schedule policy: %v", err)
		return err
	}

	fmt.Printf("Setting schedule policy %s for instance %s", policyName, vm.Name)

	if err = waitOperator(ctx, op); err != nil {
		log.Logger.Errorf("failed to set schedule policy: %v", err)
		return err
	}

	return nil
}

// UnsetSchedulePolicy detaches a schedule policy from a Google Compute Engine instance.
// It removes the specified policy from the instance's list of attached policies.
func UnsetSchedulePolicy(vm *config.VM, policyName string) error {
	ctx := context.Background()
	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Logger.Errorf("Failed to create Instances client: %v", err)
		return err
	}
	defer client.Close()

	instance, err := getInstance(ctx, vm.Project, vm.Zone, vm.Name)
	if err != nil {
		log.Logger.Errorf("Failed to get instance: %v", err)
		return err
	}
	region, err := getRegionFromInstance(instance)
	if err != nil {
		log.Logger.Errorf("Failed to get region from instance: %v", err)
		return err
	}

	policySelfLink := fmt.Sprintf("projects/%s/regions/%s/resourcePolicies/%s", vm.Project, region, policyName)

	req := &computepb.RemoveResourcePoliciesInstanceRequest{
		Instance: vm.Name,
		Project:  vm.Project,
		Zone:     vm.Zone,
		InstancesRemoveResourcePoliciesRequestResource: &computepb.InstancesRemoveResourcePoliciesRequest{
			ResourcePolicies: []string{policySelfLink},
		},
	}

	op, err := client.RemoveResourcePolicies(ctx, req)
	if err != nil {
		log.Logger.Errorf("Failed to unset schedule policy: %v", err)
		return err
	}
	fmt.Printf("Remove schedule policy %s for instance %s", policyName, vm.Name)

	if err = waitOperator(ctx, op); err != nil {
		log.Logger.Errorf("failed to unset schedule policy: %v", err)
		return err
	}

	return nil
}

func toPtr(s string) *string {
	return &s
}

// waitOperator waits for the operation to complete and prints a dot every second until the operation is done.
// It returns an error if the operation fails or if the context is canceled.
func waitOperator(ctx context.Context, op *compute.Operation) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	eg, ctx := errgroup.WithContext(ctx)
	done := make(chan struct{})
	eg.Go(func() error {
		// Wait for the operation to complete
		if err := op.Wait(ctx); err != nil {
			return err
		}
		close(done)
		return nil
	})
	eg.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done(): // Context canceled, exit the goroutine
				fmt.Println() // Print newline for clean output
				return ctx.Err()
			case <-done: // Operation is done, exit the goroutine
				fmt.Println() // Print newline for clean output
				return nil
			case <-ticker.C: // One second has passed
				fmt.Print(".")
			}
		}
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to wait for operation: %v", err)
	}
	return nil
}
