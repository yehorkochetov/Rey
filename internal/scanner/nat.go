package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

type NATGatewayScanner struct{}

func (n *NATGatewayScanner) Name() string {
	return "nat-gateway"
}

func (n *NATGatewayScanner) Scan(ctx context.Context, cfg aws.Config, t config.Thresholds) ([]DeadResource, error) {
	ec2Client := ec2.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	out, err := ec2Client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{
		Filter: []ec2types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describe nat gateways: %w", err)
	}

	now := time.Now().UTC()
	window := time.Duration(t.NATIdleDays) * 24 * time.Hour
	start := now.Add(-window)

	var results []DeadResource
	for _, ng := range out.NatGateways {
		id := aws.ToString(ng.NatGatewayId)

		if t.NATIdleDays > 0 {
			bytes, err := natBytesOut(ctx, cwClient, id, start, now, window)
			if err != nil {
				return nil, err
			}
			if bytes > 0 {
				continue
			}
		}

		tags := make(map[string]string)
		var name string
		for _, t := range ng.Tags {
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
			Type:        "NATGateway",
			ID:          id,
			Name:        name,
			Region:      cfg.Region,
			MonthlyCost: 32.40,
			Reason:      idleReason("No traffic processed", t.NATIdleDays),
			Tags:        tags,
		})
	}

	return results, nil
}

func (n *NATGatewayScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func natBytesOut(ctx context.Context, client *cloudwatch.Client, natID string, start, end time.Time, window time.Duration) (float64, error) {
	stats, err := client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/NATGateway"),
		MetricName: aws.String("BytesOutToDestination"),
		Dimensions: []cwtypes.Dimension{
			{Name: aws.String("NatGatewayId"), Value: aws.String(natID)},
		},
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		Period:     aws.Int32(int32(window.Seconds())),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
	})
	if err != nil {
		return 0, fmt.Errorf("nat metrics %s: %w", natID, err)
	}

	var total float64
	for _, dp := range stats.Datapoints {
		if dp.Sum != nil {
			total += *dp.Sum
		}
	}
	return total, nil
}
