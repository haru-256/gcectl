package gcp

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
)

type instancesClient interface {
	Get(context.Context, *computepb.GetInstanceRequest, ...gax.CallOption) (*computepb.Instance, error)
	Start(context.Context, *computepb.StartInstanceRequest, ...gax.CallOption) (*compute.Operation, error)
	Stop(context.Context, *computepb.StopInstanceRequest, ...gax.CallOption) (*compute.Operation, error)
	AddResourcePolicies(context.Context, *computepb.AddResourcePoliciesInstanceRequest, ...gax.CallOption) (*compute.Operation, error)
	RemoveResourcePolicies(context.Context, *computepb.RemoveResourcePoliciesInstanceRequest, ...gax.CallOption) (*compute.Operation, error)
	SetMachineType(context.Context, *computepb.SetMachineTypeInstanceRequest, ...gax.CallOption) (*compute.Operation, error)
	Close() error
}

type resourcePoliciesClient interface {
	Get(context.Context, *computepb.GetResourcePolicyRequest, ...gax.CallOption) (*computepb.ResourcePolicy, error)
	Close() error
}

// VMRepository implements the repository.VMRepository interface for GCP.
//
//nolint:govet // Field order optimized for readability over memory alignment
type VMRepository struct {
	logger log.Logger

	instancesClient        instancesClient
	resourcePoliciesClient resourcePoliciesClient
}

// NewVMRepository creates a VMRepository with GCP clients initialized from ctx.
// The returned repository owns the clients and must be closed by the caller.
func NewVMRepository(ctx context.Context, logger log.Logger) (*VMRepository, error) {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Instances client: %w", err)
	}

	resourcePoliciesClient, err := compute.NewResourcePoliciesRESTClient(ctx)
	if err != nil {
		if closeErr := instancesClient.Close(); closeErr != nil {
			logger.Errorf("Failed to close Instances client after ResourcePolicies client creation failed: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to create ResourcePolicies client: %w", err)
	}

	return newVMRepository(logger, instancesClient, resourcePoliciesClient), nil
}

// newVMRepository allows tests to inject GCP clients.
func newVMRepository(logger log.Logger, instancesClient instancesClient, resourcePoliciesClient resourcePoliciesClient) *VMRepository {
	return &VMRepository{
		logger:                 logger,
		instancesClient:        instancesClient,
		resourcePoliciesClient: resourcePoliciesClient,
	}
}

// Close releases any resources held by the repository, including GCP clients.
func (r *VMRepository) Close() error {
	var closeErrs []error
	if err := r.instancesClient.Close(); err != nil {
		r.logger.Errorf("Failed to close Instances client: %v", err)
		closeErrs = append(closeErrs, err)
	}
	if err := r.resourcePoliciesClient.Close(); err != nil {
		r.logger.Errorf("Failed to close ResourcePolicies client: %v", err)
		closeErrs = append(closeErrs, err)
	}
	return errors.Join(closeErrs...)
}

func (r *VMRepository) FindByName(ctx context.Context, vm *model.VM) (*model.VM, error) {
	req := &computepb.GetInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	instance, err := r.instancesClient.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	return r.toModel(ctx, instance)
}

func (r *VMRepository) Start(ctx context.Context, vm *model.VM) error {
	req := &computepb.StartInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	op, err := r.instancesClient.Start(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return r.waitOperator(ctx, op)
}

func (r *VMRepository) Stop(ctx context.Context, vm *model.VM) error {
	req := &computepb.StopInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	op, err := r.instancesClient.Stop(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return r.waitOperator(ctx, op)
}

// SetSchedulePolicy attaches a schedule policy to a Google Compute Engine instance.
func (r *VMRepository) SetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
	// Get instance details
	req := &computepb.GetInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	instance, err := r.instancesClient.Get(ctx, req)
	if err != nil {
		r.logger.Errorf("failed to get instance: %v", err)
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Extract region from zone
	region, err := extractRegion(instance.GetZone())
	if err != nil {
		r.logger.Errorf("Failed to get region from instance: %v", err)
		return fmt.Errorf("failed to extract region: %w", err)
	}

	policySelfLink := fmt.Sprintf("projects/%s/regions/%s/resourcePolicies/%s", vm.Project, region, policyName)

	addPolicyReq := &computepb.AddResourcePoliciesInstanceRequest{
		Instance: vm.Name,
		Project:  vm.Project,
		Zone:     vm.Zone,
		InstancesAddResourcePoliciesRequestResource: &computepb.InstancesAddResourcePoliciesRequest{
			ResourcePolicies: []string{policySelfLink},
		},
	}

	op, err := r.instancesClient.AddResourcePolicies(ctx, addPolicyReq)
	if err != nil {
		r.logger.Errorf("Failed to set schedule policy: %v", err)
		return fmt.Errorf("failed to add resource policy: %w", err)
	}

	r.logger.Infof("Setting schedule policy %s for instance %s", policyName, vm.Name)

	if err = r.waitOperator(ctx, op); err != nil {
		r.logger.Errorf("failed to wait for operation: %v", err)
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UnsetSchedulePolicy removes a schedule policy from a Google Compute Engine instance.
func (r *VMRepository) UnsetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
	// Get instance details
	req := &computepb.GetInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	instance, err := r.instancesClient.Get(ctx, req)
	if err != nil {
		r.logger.Errorf("failed to get instance: %v", err)
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Extract region from zone
	region, err := extractRegion(instance.GetZone())
	if err != nil {
		r.logger.Errorf("Failed to get region from instance: %v", err)
		return fmt.Errorf("failed to extract region: %w", err)
	}

	policySelfLink := fmt.Sprintf("projects/%s/regions/%s/resourcePolicies/%s", vm.Project, region, policyName)

	removePolicyReq := &computepb.RemoveResourcePoliciesInstanceRequest{
		Instance: vm.Name,
		Project:  vm.Project,
		Zone:     vm.Zone,
		InstancesRemoveResourcePoliciesRequestResource: &computepb.InstancesRemoveResourcePoliciesRequest{
			ResourcePolicies: []string{policySelfLink},
		},
	}

	op, err := r.instancesClient.RemoveResourcePolicies(ctx, removePolicyReq)
	if err != nil {
		r.logger.Errorf("Failed to unset schedule policy: %v", err)
		return fmt.Errorf("failed to remove resource policy: %w", err)
	}

	r.logger.Infof("Removing schedule policy %s from instance %s", policyName, vm.Name)

	if err = r.waitOperator(ctx, op); err != nil {
		r.logger.Errorf("failed to wait for operation: %v", err)
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateMachineType changes the machine type of a VM instance.
func (r *VMRepository) UpdateMachineType(ctx context.Context, vm *model.VM, machineType string) error {
	// Machine type must be in the format: zones/ZONE/machineTypes/MACHINE_TYPE
	machineTypeURL := fmt.Sprintf("zones/%s/machineTypes/%s", vm.Zone, machineType)

	setMachineTypeReq := &computepb.SetMachineTypeInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
		InstancesSetMachineTypeRequestResource: &computepb.InstancesSetMachineTypeRequest{
			MachineType: &machineTypeURL,
		},
	}

	op, err := r.instancesClient.SetMachineType(ctx, setMachineTypeReq)
	if err != nil {
		r.logger.Errorf("Failed to set machine type: %v", err)
		return fmt.Errorf("failed to set machine type: %w", err)
	}

	r.logger.Infof("Setting machine type to %s for instance %s", machineType, vm.Name)

	if err = r.waitOperator(ctx, op); err != nil {
		r.logger.Errorf("failed to wait for operation: %v", err)
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// toModel converts a GCP instance to domain model
func (r *VMRepository) toModel(ctx context.Context, instance *computepb.Instance) (*model.VM, error) {
	vm := &model.VM{
		Name:        instance.GetName(),
		Status:      model.StatusFromString(instance.GetStatus()),
		MachineType: extractMachineType(instance.GetMachineType()),
	}

	// Extract project and zone from instance
	project, err := extractProject(instance.GetSelfLink())
	if err != nil {
		return nil, fmt.Errorf("failed to extract project from instance: %w", err)
	}
	zone, err := extractZone(instance.GetZone())
	if err != nil {
		return nil, fmt.Errorf("failed to extract zone from instance: %w", err)
	}
	vm.Project = project
	vm.Zone = zone

	// Parse start time
	if startTimeStr := instance.GetLastStartTimestamp(); startTimeStr != "" {
		if startTime, parseErr := time.Parse(time.RFC3339, startTimeStr); parseErr == nil {
			vm.LastStartTime = &startTime
		}
	}

	// Get schedule policy (existing logic)
	r.logger.Debugf("Getting schedule policy for instance %s", vm.Name)
	schedulePolicy, err := r.getSchedulePolicy(ctx, instance)
	if err != nil {
		r.logger.Errorf("Failed to get schedule policy: %v", err)
		return nil, err
	}
	vm.SchedulePolicy = schedulePolicy

	return vm, nil
}

func (r *VMRepository) getSchedulePolicy(ctx context.Context, instance *computepb.Instance) (string, error) {
	policies := instance.GetResourcePolicies()
	if len(policies) == 0 {
		return "", nil
	}

	project, err := extractProject(instance.GetSelfLink())
	if err != nil {
		r.logger.Errorf("Failed to get project from instance: %v", err)
		return "", err
	}

	region, err := extractRegion(instance.GetZone())
	if err != nil {
		r.logger.Errorf("Failed to get region from instance: %v", err)
		return "", err
	}

	// 順次処理（ポリシー数は通常少ないため）
	for _, policy := range policies {
		r.logger.Debugf("Resource Policy: %s", policy)

		policyParts := strings.Split(policy, "/")
		policyName := policyParts[len(policyParts)-1]

		policyReq := &computepb.GetResourcePolicyRequest{
			Project:        project,
			Region:         region,
			ResourcePolicy: policyName,
		}

		var resourcePolicy *computepb.ResourcePolicy
		resourcePolicy, err = r.resourcePoliciesClient.Get(ctx, policyReq)
		if err != nil {
			r.logger.Errorf("Failed to get resource policy details: %v", err)
			continue
		}

		schedulePolicy := resourcePolicy.GetInstanceSchedulePolicy()
		if formattedPolicy := formatInstanceSchedulePolicy(policyName, schedulePolicy); formattedPolicy != "" {
			return formattedPolicy, nil
		}
	}

	return "", nil
}

func formatInstanceSchedulePolicy(policyName string, schedulePolicy *computepb.ResourcePolicyInstanceSchedulePolicy) string {
	if schedulePolicy == nil {
		return ""
	}
	if schedulePolicy.VmStopSchedule == nil {
		return policyName
	}
	schedule := schedulePolicy.VmStopSchedule.GetSchedule()
	if schedule == "" {
		return policyName
	}
	return fmt.Sprintf("%s(%s)", policyName, schedule)
}

func extractMachineType(fullURI string) string {
	pattern := `machineTypes/([^/]+)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(fullURI)
	if len(matches) < 2 {
		return "UNKNOWN"
	}
	return matches[1]
}

func extractProject(selfLink string) (string, error) {
	pattern := `projects/([^/]+)/`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(selfLink)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract project")
	}
	return matches[1], nil
}

func extractZone(zoneURI string) (string, error) {
	parts := strings.Split(zoneURI, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid zone URI")
	}
	return parts[len(parts)-1], nil
}

// extractRegion extracts the region from a zone URI
// Example: "https://www.googleapis.com/compute/v1/projects/PROJECT/zones/us-central1-a" -> "us-central1"
func extractRegion(zoneURI string) (string, error) {
	parts := strings.Split(zoneURI, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid zone URI")
	}

	zoneName := parts[len(parts)-1]
	// Zone format: "region-zone" (e.g., "us-central1-a")
	// Extract region by removing the last part after the last hyphen
	lastHyphen := strings.LastIndex(zoneName, "-")
	if lastHyphen == -1 {
		return "", fmt.Errorf("invalid zone format: %s", zoneName)
	}

	return zoneName[:lastHyphen], nil
}

// waitOperator waits for the operation to complete and optionally reports progress.
//
// This method monitors a GCP compute operation until completion. If a progress callback
// has been set via SetProgressCallback(), it will be called every second during the wait.
// This allows the presentation layer to display progress (e.g., dots) without violating
// Clean Architecture principles.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - op: The GCP compute operation to wait for
//
// Returns:
//   - error: Error if the operation fails or context is canceled
//
// Example:
//
//	repo.SetProgressCallback(console.Progress)
//	err := repo.waitOperator(ctx, operation)
func (r *VMRepository) waitOperator(ctx context.Context, op *compute.Operation) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	return op.Wait(ctx)
}

var _ repository.VMRepository = (*VMRepository)(nil)
