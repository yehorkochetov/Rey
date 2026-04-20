package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

type EBSScanner struct{}

func (e *EBSScanner) Name() string {
	return "ebs-volume"
}

func (e *EBSScanner) Scan(ctx context.Context, cfg aws.Config, t config.Thresholds) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	out, err := client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("status"),
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describe volumes: %w", err)
	}

	now := time.Now().UTC()
	var results []DeadResource
	for _, v := range out.Volumes {
		if r, ok := considerEBSVolume(v, now, t, cfg.Region); ok {
			results = append(results, r)
		}
	}
	return results, nil
}

func (e *EBSScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

// considerEBSVolume is the pure decision helper used by EBSScanner. It takes
// an already-filtered "available" volume, the current time, and the active
// thresholds, and returns the DeadResource plus whether the volume should be
// flagged. A threshold of 0 disables the age check entirely.
func considerEBSVolume(v ec2types.Volume, now time.Time, t config.Thresholds, region string) (DeadResource, bool) {
	var age time.Duration
	if v.CreateTime != nil {
		age = now.Sub(*v.CreateTime)
	}
	if t.EBSUnattachedDays > 0 {
		minAge := time.Duration(t.EBSUnattachedDays) * 24 * time.Hour
		if age < minAge {
			return DeadResource{}, false
		}
	}

	tags := make(map[string]string)
	var name string
	for _, tag := range v.Tags {
		k := aws.ToString(tag.Key)
		val := aws.ToString(tag.Value)
		tags[k] = val
		if k == "Name" {
			name = val
		}
	}
	id := aws.ToString(v.VolumeId)
	if name == "" {
		name = id
	}
	var size int32
	if v.Size != nil {
		size = *v.Size
	}

	reason := "Unattached"
	if age > 0 {
		reason = fmt.Sprintf("Unattached for %d days", int(age.Hours()/24))
	}

	return DeadResource{
		Type:        "EBSVolume",
		ID:          id,
		Name:        name,
		Region:      region,
		Age:         age,
		MonthlyCost: float64(size) * 0.10,
		Reason:      reason,
		Tags:        tags,
	}, true
}
