package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

type VPCEndpointScanner struct{}

func (v *VPCEndpointScanner) Name() string {
	return "vpc-endpoint"
}

func (v *VPCEndpointScanner) Scan(ctx context.Context, cfg aws.Config, _ config.Thresholds) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	out, err := client.DescribeVpcEndpoints(ctx, &ec2.DescribeVpcEndpointsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("vpc-endpoint-state"),
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describe vpc endpoints: %w", err)
	}

	var results []DeadResource
	for _, ep := range out.VpcEndpoints {
		if vpcEndpointInUse(ep) {
			continue
		}

		id := aws.ToString(ep.VpcEndpointId)
		tags := make(map[string]string)
		var name string
		for _, t := range ep.Tags {
			k := aws.ToString(t.Key)
			val := aws.ToString(t.Value)
			tags[k] = val
			if k == "Name" {
				name = val
			}
		}
		if name == "" {
			name = id
		}

		results = append(results, DeadResource{
			Type:        "VPCEndpoint",
			ID:          id,
			Name:        name,
			Region:      cfg.Region,
			MonthlyCost: vpcEndpointCost(ep.VpcEndpointType),
			Reason:      "No route tables or security groups associated",
			Tags:        tags,
		})
	}

	return results, nil
}

func (v *VPCEndpointScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func vpcEndpointInUse(ep ec2types.VpcEndpoint) bool {
	switch ep.VpcEndpointType {
	case ec2types.VpcEndpointTypeGateway:
		return len(ep.RouteTableIds) > 0
	case ec2types.VpcEndpointTypeInterface, ec2types.VpcEndpointTypeGatewayLoadBalancer:
		return len(ep.Groups) > 0
	default:
		return len(ep.RouteTableIds) > 0 || len(ep.Groups) > 0
	}
}

func vpcEndpointCost(t ec2types.VpcEndpointType) float64 {
	if t == ec2types.VpcEndpointTypeGateway {
		return 0
	}
	return 7.20
}
