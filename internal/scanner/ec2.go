package scanner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2Scanner struct {
	MinAge time.Duration
}

func (e *EC2Scanner) Name() string {
	return "ec2-stopped"
}

func (e *EC2Scanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	out, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"stopped"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describe instances: %w", err)
	}

	var results []DeadResource
	now := time.Now().UTC()
	for _, res := range out.Reservations {
		for _, inst := range res.Instances {
			stopTime, ok := parseStopTime(aws.ToString(inst.StateTransitionReason))
			if !ok {
				continue
			}
			age := now.Sub(stopTime)
			if age < e.MinAge {
				continue
			}

			tags := make(map[string]string)
			var name string
			for _, t := range inst.Tags {
				k := aws.ToString(t.Key)
				v := aws.ToString(t.Value)
				tags[k] = v
				if k == "Name" {
					name = v
				}
			}
			id := aws.ToString(inst.InstanceId)
			if name == "" {
				name = id
			}
			days := int(age.Hours() / 24)

			results = append(results, DeadResource{
				Type:        "EC2Instance",
				ID:          id,
				Name:        name,
				Region:      cfg.Region,
				Age:         age,
				MonthlyCost: 0,
				Reason:      fmt.Sprintf("Stopped for %d days, attached EBS still charging", days),
				Tags:        tags,
			})
		}
	}

	return results, nil
}

func (e *EC2Scanner) EstimateCost(r DeadResource) float64 {
	return 0
}

// parseStopTime extracts the timestamp from an EC2 StateTransitionReason
// string like "User initiated (2024-01-15 10:30:45 GMT)".
func parseStopTime(reason string) (time.Time, bool) {
	open := strings.Index(reason, "(")
	end := strings.Index(reason, " GMT)")
	if open < 0 || end < 0 || end <= open+1 {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02 15:04:05", reason[open+1:end])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
