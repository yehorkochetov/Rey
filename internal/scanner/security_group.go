package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

type SecurityGroupScanner struct{}

func (s *SecurityGroupScanner) Name() string {
	return "security-group"
}

func (s *SecurityGroupScanner) Scan(ctx context.Context, cfg aws.Config, _ config.Thresholds) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	groups, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("describe security groups: %w", err)
	}

	enis, err := client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{})
	if err != nil {
		return nil, fmt.Errorf("describe network interfaces: %w", err)
	}

	inUse := make(map[string]struct{})
	for _, eni := range enis.NetworkInterfaces {
		for _, g := range eni.Groups {
			inUse[aws.ToString(g.GroupId)] = struct{}{}
		}
	}

	var results []DeadResource
	for _, g := range groups.SecurityGroups {
		if r, ok := considerSecurityGroup(g, inUse, cfg.Region); ok {
			results = append(results, r)
		}
	}
	return results, nil
}

func (s *SecurityGroupScanner) EstimateCost(r DeadResource) float64 {
	return 0
}

// considerSecurityGroup is the pure filter for security groups. The
// "default" group is untouched — AWS creates one per VPC and it can't be
// deleted. Any other group is flagged when no ENI references it.
func considerSecurityGroup(g ec2types.SecurityGroup, inUse map[string]struct{}, region string) (DeadResource, bool) {
	name := aws.ToString(g.GroupName)
	if name == "default" {
		return DeadResource{}, false
	}
	id := aws.ToString(g.GroupId)
	if _, used := inUse[id]; used {
		return DeadResource{}, false
	}

	tags := make(map[string]string)
	for _, tag := range g.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	return DeadResource{
		Type:        "SecurityGroup",
		ID:          id,
		Name:        name,
		Region:      region,
		MonthlyCost: 0,
		Reason:      "Not attached to any resource",
		Tags:        tags,
	}, true
}
