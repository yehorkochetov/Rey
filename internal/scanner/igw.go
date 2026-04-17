package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type IGWScanner struct{}

func (i *IGWScanner) Name() string {
	return "internet-gateway"
}

func (i *IGWScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	out, err := client.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("describe internet gateways: %w", err)
	}

	var results []DeadResource
	for _, igw := range out.InternetGateways {
		if hasActiveAttachment(igw.Attachments) {
			continue
		}

		id := aws.ToString(igw.InternetGatewayId)
		tags := make(map[string]string)
		var name string
		for _, t := range igw.Tags {
			k := aws.ToString(t.Key)
			v := aws.ToString(t.Value)
			tags[k] = v
			if k == "Name" {
				name = v
			}
		}
		if name == "" {
			name = id
		}

		results = append(results, DeadResource{
			Type:        "InternetGateway",
			ID:          id,
			Name:        name,
			Region:      cfg.Region,
			MonthlyCost: 0,
			Reason:      "Not attached to any VPC",
			Tags:        tags,
		})
	}

	return results, nil
}

func (i *IGWScanner) EstimateCost(r DeadResource) float64 {
	return 0
}

func hasActiveAttachment(attachments []ec2types.InternetGatewayAttachment) bool {
	for _, a := range attachments {
		if a.State != ec2types.AttachmentStatusDetached {
			return true
		}
	}
	return false
}
