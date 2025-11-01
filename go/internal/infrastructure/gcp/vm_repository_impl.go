package gcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"golang.org/x/sync/errgroup"

	"github.com/haru-256/gcectl/internal/domain/model"
	"github.com/haru-256/gcectl/internal/domain/repository"
	"github.com/haru-256/gcectl/internal/infrastructure/config"
	"github.com/haru-256/gcectl/internal/infrastructure/log"
)

// ProgressCallback is a function type for reporting operation progress.
//
// This callback is invoked periodically (approximately once per second) while waiting
// for long-running GCP operations to complete. It allows the presentation layer to
// display progress indicators (e.g., dots, spinner) without coupling the infrastructure
// layer to specific output mechanisms.
//
// The callback takes no parameters and returns no values. It should be a lightweight
// operation, typically just printing a character or updating a progress indicator.
//
// Example:
//
//	repo.SetProgressCallback(func() {
//	    fmt.Print(".")
//	})
type ProgressCallback func()

// VMRepository implements the repository.VMRepository interface for GCP.
//
//nolint:govet // Field order optimized for readability over memory alignment
type VMRepository struct {
	configPath       string
	logger           log.Logger
	progressCallback ProgressCallback // Optional callback for operation progress
}

// NewVMRepository creates a new VMRepository instance.
//
// Parameters:
//   - configPath: Path to the configuration file
//   - logger: Logger instance for logging
//
// Returns:
//   - *VMRepository: A new repository instance
func NewVMRepository(configPath string, logger log.Logger) *VMRepository {
	return &VMRepository{
		configPath: configPath,
		logger:     logger,
	}
}

// SetProgressCallback sets a callback function to be called during operation progress.
//
// This method allows the presentation layer to display progress (e.g., dots) during
// long-running GCP operations without violating Clean Architecture principles.
// The callback will be invoked approximately once per second while waiting for
// operations to complete.
//
// Parameters:
//   - callback: Function to call periodically during operations
//
// Example:
//
//	repo := gcp.NewVMRepository(configPath, logger)
//	repo.SetProgressCallback(console.Progress)
//	repo.Start(ctx, vm) // Will call console.Progress() periodically
func (r *VMRepository) SetProgressCallback(callback ProgressCallback) {
	r.progressCallback = callback
}

func (r *VMRepository) FindByName(ctx context.Context, vm *model.VM) (*model.VM, error) {
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			r.logger.Errorf("Failed to close client: %v", closeErr)
		}
	}()

	req := &computepb.GetInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	instance, err := client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	return r.toModel(ctx, instance)
}

func (r *VMRepository) FindAll(ctx context.Context) ([]*model.VM, error) {
	// 設定ファイルから VM リストを読み込み
	cfg, err := config.ParseConfig(r.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// errgroup を使用して並行実行
	eg, ctx := errgroup.WithContext(ctx)
	vmChan := make(chan *model.VM, len(cfg.VMs))

	for _, cfgVM := range cfg.VMs {
		cfgVM := cfgVM // ループ変数のキャプチャ
		eg.Go(func() error {
			vm, findErr := r.FindByName(ctx, cfgVM)
			if findErr != nil {
				// エラーをログに記録して続行
				r.logger.Errorf("failed to find VM %s in project %s zone %s: %v", cfgVM.Name, cfgVM.Project, cfgVM.Zone, findErr)
				return nil // エラーを返さずに続行
			}
			vmChan <- vm
			return nil
		})
	}

	// すべてのゴルーチンが完了するのを待つ
	if waitErr := eg.Wait(); waitErr != nil {
		return nil, fmt.Errorf("failed to fetch VMs: %w", waitErr)
	}
	close(vmChan)

	// チャネルから結果を収集
	vms := make([]*model.VM, 0, len(cfg.VMs))
	for vm := range vmChan {
		vms = append(vms, vm)
	}

	return vms, nil
}

func (r *VMRepository) Start(ctx context.Context, vm *model.VM) error {
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			r.logger.Errorf("Failed to close client: %v", closeErr)
		}
	}()

	req := &computepb.StartInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	op, err := client.Start(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return r.waitOperator(ctx, op)
}

func (r *VMRepository) Stop(ctx context.Context, vm *model.VM) error {
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			r.logger.Errorf("Failed to close client: %v", closeErr)
		}
	}()

	req := &computepb.StopInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	op, err := client.Stop(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return r.waitOperator(ctx, op)
}

// SetSchedulePolicy attaches a schedule policy to a Google Compute Engine instance.
func (r *VMRepository) SetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		r.logger.Errorf("failed to create Instances client: %v", err)
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			r.logger.Errorf("Failed to close client: %v", closeErr)
		}
	}()

	// Get instance details
	req := &computepb.GetInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	instance, err := client.Get(ctx, req)
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

	op, err := client.AddResourcePolicies(ctx, addPolicyReq)
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
	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		r.logger.Errorf("failed to create Instances client: %v", err)
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			r.logger.Errorf("Failed to close client: %v", closeErr)
		}
	}()

	// Get instance details
	req := &computepb.GetInstanceRequest{
		Project:  vm.Project,
		Zone:     vm.Zone,
		Instance: vm.Name,
	}

	instance, err := client.Get(ctx, req)
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

	op, err := client.RemoveResourcePolicies(ctx, removePolicyReq)
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
	// Create a new InstancesClient with authentication
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		r.logger.Errorf("failed to create Instances client: %v", err)
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			r.logger.Errorf("Failed to close client: %v", closeErr)
		}
	}()

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

	op, err := client.SetMachineType(ctx, setMachineTypeReq)
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
	defaultPolicy := "#NONE"

	policies := instance.GetResourcePolicies()
	if len(policies) == 0 {
		return defaultPolicy, nil
	}

	policyClient, err := compute.NewResourcePoliciesRESTClient(ctx)
	if err != nil {
		r.logger.Errorf("Failed to create ResourcePolicies client: %v", err)
		return "", err
	}
	defer func() {
		if closeErr := policyClient.Close(); closeErr != nil {
			r.logger.Errorf("Failed to close policy client: %v", closeErr)
		}
	}()

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
		resourcePolicy, err = policyClient.Get(ctx, policyReq)
		if err != nil {
			r.logger.Errorf("Failed to get resource policy details: %v", err)
			continue
		}

		schedulePolicy := resourcePolicy.GetInstanceSchedulePolicy()
		if schedulePolicy != nil {
			return fmt.Sprintf("%s(%s)", policyName, *schedulePolicy.VmStopSchedule.Schedule), nil
		}
	}

	return defaultPolicy, nil
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

	// Only start progress reporting if callback is set
	if r.progressCallback != nil {
		eg.Go(func() error {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done(): // Context canceled, exit the goroutine
					return ctx.Err()
				case <-done: // Operation is done, exit the goroutine
					return nil
				case <-ticker.C: // One second has passed
					r.progressCallback()
				}
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to wait for operation: %v", err)
	}
	return nil
}

var _ repository.VMRepository = (*VMRepository)(nil)
