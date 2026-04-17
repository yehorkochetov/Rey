package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type ENIScanner struct{}

func (e *ENIScanner) Name() string {
	return "network-interface"
}

func (e *ENIScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	out, err := client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("status"),
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describe network interfaces: %w", err)
	}

	var results []DeadResource
	for _, eni := range out.NetworkInterfaces {
		id := aws.ToString(eni.NetworkInterfaceId)
		name := aws.ToString(eni.Description)
		if name == "" {
			name = id
		}

		tags := make(map[string]string)
		for _, t := range eni.TagSet {
			tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
		}

		results = append(results, DeadResource{
			Type:        "NetworkInterface",
			ID:          id,
			Name:        name,
			Region:      cfg.Region,
			MonthlyCost: 0,
			Reason:      "Network interface not attached to any resource",
			Tags:        tags,
		})
	}

	return results, nil
}

func (e *ENIScanner) EstimateCost(r DeadResource) float64 {
	return 0
}
