package gcp

import (
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
)

func TestFormatInstanceSchedulePolicy(t *testing.T) {
	tests := []struct {
		name       string
		policyName string
		policy     *computepb.ResourcePolicyInstanceSchedulePolicy
		want       string
	}{
		{
			name:       "nil policy returns empty string",
			policyName: "nightly-stop",
			policy:     nil,
			want:       "",
		},
		{
			name:       "nil stop schedule returns policy name",
			policyName: "nightly-stop",
			policy:     &computepb.ResourcePolicyInstanceSchedulePolicy{},
			want:       "nightly-stop",
		},
		{
			name:       "nil schedule returns policy name",
			policyName: "nightly-stop",
			policy: &computepb.ResourcePolicyInstanceSchedulePolicy{
				VmStopSchedule: &computepb.ResourcePolicyInstanceSchedulePolicySchedule{},
			},
			want: "nightly-stop",
		},
		{
			name:       "schedule returns policy name with schedule",
			policyName: "nightly-stop",
			policy: &computepb.ResourcePolicyInstanceSchedulePolicy{
				VmStopSchedule: &computepb.ResourcePolicyInstanceSchedulePolicySchedule{Schedule: stringPtr("0 22 * * *")},
			},
			want: "nightly-stop(0 22 * * *)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatInstanceSchedulePolicy(tt.policyName, tt.policy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func stringPtr(v string) *string {
	return &v
}
