package graph

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	// TODO update back to flux-sched when merged
)

// generateTopologyInput generates the parameters for the topology request
// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeInstanceTopologyInput
func generateTopologyInput(group string, instance string) *ec2.DescribeInstanceTopologyInput {

	groups := []*string{}
	instances := []*string{}
	dryRun := false

	// For larger sets we might want NextToken (string) or Filters []*ec2Filter{}
	input := ec2.DescribeInstanceTopologyInput{
		DryRun: &dryRun,
	}

	// Don't add these empty if not provided, likely weird errors
	if group != "" {
		groups = append(groups, &group)
		input.GroupNames = groups
	}
	if instance != "" {
		instances = append(instances, &instance)
		input.InstanceIds = instances
	}
	return &input
}
