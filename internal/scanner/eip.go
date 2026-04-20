package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/yehorkochetov/rey/internal/config"
)

type EIPScanner struct{}

func (e *EIPScanner) Name() string {
	return "elastic-ip"
}

func (e *EIPScanner) Scan(ctx context.Context, cfg aws.Config, _ config.Thresholds) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	out, err := client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		return nil, fmt.Errorf("describe addresses: %w", err)
	}

	var results []DeadResource
	for _, addr := range out.Addresses {
		if addr.AssociationId != nil {
			continue
		}

		tags := make(map[string]string)
		for _, t := range addr.Tags {
			tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
		}

		results = append(results, DeadResource{
			Type:        "ElasticIP",
			ID:          aws.ToString(addr.AllocationId),
			Name:        aws.ToString(addr.PublicIp),
			Region:      cfg.Region,
			Reason:      "Not associated to any instance or ENI",
			MonthlyCost: 3.60,
			Tags:        tags,
		})
	}

	return results, nil
}

func (e *EIPScanner) EstimateCost(r DeadResource) float64 {
	return 3.60
}
