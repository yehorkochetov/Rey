package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type SecurityGroupScanner struct{}

func (s *SecurityGroupScanner) Name() string {
	return "security-group"
}

func (s *SecurityGroupScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
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
		name := aws.ToString(g.GroupName)
		if name == "default" {
			continue
		}
		id := aws.ToString(g.GroupId)
		if _, used := inUse[id]; used {
			continue
		}

		tags := make(map[string]string)
		for _, t := range g.Tags {
			tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
		}

		results = append(results, DeadResource{
			Type:        "SecurityGroup",
			ID:          id,
			Name:        name,
			Region:      cfg.Region,
			MonthlyCost: 0,
			Reason:      "Not attached to any resource",
			Tags:        tags,
		})
	}

	return results, nil
}

func (s *SecurityGroupScanner) EstimateCost(r DeadResource) float64 {
	return 0
}
