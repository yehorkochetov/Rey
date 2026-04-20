package scanner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

type EC2Scanner struct{}

func (e *EC2Scanner) Name() string {
	return "ec2-stopped"
}

func (e *EC2Scanner) Scan(ctx context.Context, cfg aws.Config, t config.Thresholds) ([]DeadResource, error) {
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

	now := time.Now().UTC()
	var results []DeadResource
	for _, res := range out.Reservations {
		for _, inst := range res.Instances {
			if r, ok := considerEC2Instance(inst, now, t, cfg.Region); ok {
				results = append(results, r)
			}
		}
	}
	return results, nil
}

// considerEC2Instance decides whether a single instance should be flagged.
// Only stopped instances with a parseable stop timestamp qualify; any other
// state is skipped even if callers forget to apply the API-level filter.
func considerEC2Instance(inst ec2types.Instance, now time.Time, t config.Thresholds, region string) (DeadResource, bool) {
	if inst.State == nil || inst.State.Name != ec2types.InstanceStateNameStopped {
		return DeadResource{}, false
	}
	stopTime, ok := parseStopTime(aws.ToString(inst.StateTransitionReason))
	if !ok {
		return DeadResource{}, false
	}
	age := now.Sub(stopTime)
	if t.EC2StoppedDays > 0 {
		minAge := time.Duration(t.EC2StoppedDays) * 24 * time.Hour
		if age < minAge {
			return DeadResource{}, false
		}
	}

	tags := make(map[string]string)
	var name string
	for _, tag := range inst.Tags {
		k := aws.ToString(tag.Key)
		v := aws.ToString(tag.Value)
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

	return DeadResource{
		Type:        "EC2Instance",
		ID:          id,
		Name:        name,
		Region:      region,
		Age:         age,
		MonthlyCost: 0,
		Reason:      fmt.Sprintf("Stopped for %d days, attached EBS still charging", days),
		Tags:        tags,
	}, true
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
