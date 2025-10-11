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

type VMRepositoryImpl struct {
	logger     log.Logger
	configPath string
}

func NewVMRepository(configPath string, logger log.Logger) repository.VMRepository {
	return &VMRepositoryImpl{
		configPath: configPath,
		logger:     logger,
	}
}

func (r *VMRepositoryImpl) FindByName(ctx context.Context, project, zone, name string) (*model.VM, error) {
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
		Project:  project,
		Zone:     zone,
		Instance: name,
	}

	instance, err := client.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	return r.toModel(ctx, instance)
}

func (r *VMRepositoryImpl) FindAll(ctx context.Context) ([]*model.VM, error) {
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
			vm, findErr := r.FindByName(ctx, cfgVM.Project, cfgVM.Zone, cfgVM.Name)
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

func (r *VMRepositoryImpl) Start(ctx context.Context, vm *model.VM) error {
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

	return waitOperator(ctx, op)
}

func (r *VMRepositoryImpl) Stop(ctx context.Context, vm *model.VM) error {
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

	return waitOperator(ctx, op)
}

// SetSchedulePolicy attaches a schedule policy to a Google Compute Engine instance.
func (r *VMRepositoryImpl) SetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
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

	if err = waitOperator(ctx, op); err != nil {
		r.logger.Errorf("failed to wait for operation: %v", err)
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UnsetSchedulePolicy removes a schedule policy from a Google Compute Engine instance.
func (r *VMRepositoryImpl) UnsetSchedulePolicy(ctx context.Context, vm *model.VM, policyName string) error {
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

	if err = waitOperator(ctx, op); err != nil {
		r.logger.Errorf("failed to wait for operation: %v", err)
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateMachineType changes the machine type of a VM instance.
func (r *VMRepositoryImpl) UpdateMachineType(ctx context.Context, vm *model.VM, machineType string) error {
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

	if err = waitOperator(ctx, op); err != nil {
		r.logger.Errorf("failed to wait for operation: %v", err)
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// toModel converts a GCP instance to domain model
func (r *VMRepositoryImpl) toModel(ctx context.Context, instance *computepb.Instance) (*model.VM, error) {
	vm := &model.VM{
		Name:        instance.GetName(),
		Status:      model.StatusFromString(instance.GetStatus()),
		MachineType: extractMachineType(instance.GetMachineType()),
	}

	// Extract project and zone from instance
	project, _ := extractProject(instance.GetSelfLink())
	zone, _ := extractZone(instance.GetZone())
	vm.Project = project
	vm.Zone = zone

	// Parse start time
	if startTimeStr := instance.GetLastStartTimestamp(); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
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

func (r *VMRepositoryImpl) getSchedulePolicy(ctx context.Context, instance *computepb.Instance) (string, error) {
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
