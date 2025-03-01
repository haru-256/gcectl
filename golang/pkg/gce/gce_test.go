package gce

import (
	"context"
	"testing"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
)

func TestGetInstanceIntegration(t *testing.T) {
	t.Skip("Skipping integration test - requires GCP credentials")

	// Set up your test parameters
	projectID := "haru256-sandbox-20250225"
	zone := "us-central1-a"
	instanceName := "sandbox"

	// Call the function
	ctx := context.Background()
	instance, err := getInstance(ctx, projectID, zone, instanceName)

	// Check results
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

func TestGetSchedulePolicy(t *testing.T) {
	t.Skip("Skipping integration test - requires GCP credentials")

	// Set up your test parameters
	projectID := "haru256-sandbox-20250225"
	zone := "us-central1-a"
	instanceName := "sandbox"

	// Call the function
	ctx := context.Background()
	instance, err := getInstance(ctx, projectID, zone, instanceName)
	policy, err := getSchedulePolicy(ctx, instance)

	// Check results
	assert.NoError(t, err)
	assert.NotEmpty(t, policy)
}

func TestGetRegionFromInstance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		instance    *computepb.Instance
		wantRegion  string
		expectError bool
	}{
		{
			name: "valid zone us-central1-a",
			instance: &computepb.Instance{
				Zone: stringPtr("https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a"),
			},
			wantRegion:  "us-central1",
			expectError: false,
		},
		{
			name: "valid zone asia-northeast1-c",
			instance: &computepb.Instance{
				Zone: stringPtr("https://www.googleapis.com/compute/v1/projects/test-project/zones/asia-northeast1-c"),
			},
			wantRegion:  "asia-northeast1",
			expectError: false,
		},
		{
			name: "multi-dash zone europe-west3-a",
			instance: &computepb.Instance{
				Zone: stringPtr("https://www.googleapis.com/compute/v1/projects/test-project/zones/europe-west3-a"),
			},
			wantRegion:  "europe-west3",
			expectError: false,
		},
		{
			name: "missing zone",
			instance: &computepb.Instance{
				Zone: stringPtr(""),
			},
			wantRegion:  "",
			expectError: true,
		},
		{
			name: "invalid zone format",
			instance: &computepb.Instance{
				Zone: stringPtr("invalid-format"),
			},
			wantRegion:  "",
			expectError: true,
		},
		{
			name:        "nil instance",
			instance:    nil,
			wantRegion:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRegion, err := getRegionFromInstance(tt.instance)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRegion, gotRegion)
			}
		})
	}
}

func TestGetProjectFromInstance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		instance    *computepb.Instance
		wantProject string
		expectError bool
	}{
		{
			name: "valid self link",
			instance: &computepb.Instance{
				SelfLink: stringPtr("https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a/instances/instance-1"),
			},
			wantProject: "test-project",
			expectError: false,
		},
		{
			name: "valid self link with different project",
			instance: &computepb.Instance{
				SelfLink: stringPtr("https://www.googleapis.com/compute/v1/projects/other-project-123/zones/asia-northeast1-c/instances/instance-2"),
			},
			wantProject: "other-project-123",
			expectError: false,
		},
		{
			name: "valid self link with hyphens and numbers",
			instance: &computepb.Instance{
				SelfLink: stringPtr("https://www.googleapis.com/compute/v1/projects/my-project-id-42/zones/europe-west3-a/instances/instance-3"),
			},
			wantProject: "my-project-id-42",
			expectError: false,
		},
		{
			name: "empty self link",
			instance: &computepb.Instance{
				SelfLink: stringPtr(""),
			},
			wantProject: "",
			expectError: true,
		},
		{
			name: "invalid self link format",
			instance: &computepb.Instance{
				SelfLink: stringPtr("invalid-format-without-project"),
			},
			wantProject: "",
			expectError: true,
		},
		{
			name:        "nil instance",
			instance:    nil,
			wantProject: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProject, err := getProjectFromInstance(tt.instance)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantProject, gotProject)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
