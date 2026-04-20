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

type SnapshotScanner struct{}

func (s *SnapshotScanner) Name() string {
	return "ebs-snapshot"
}

func (s *SnapshotScanner) Scan(ctx context.Context, cfg aws.Config, t config.Thresholds) ([]DeadResource, error) {
	client := ec2.NewFromConfig(cfg)

	images, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	})
	if err != nil {
		return nil, fmt.Errorf("describe images: %w", err)
	}

	amiSnapshots := make(map[string]struct{})
	for _, img := range images.Images {
		for _, m := range img.BlockDeviceMappings {
			if m.Ebs == nil || m.Ebs.SnapshotId == nil {
				continue
			}
			amiSnapshots[*m.Ebs.SnapshotId] = struct{}{}
		}
	}

	snaps, err := client.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})
	if err != nil {
		return nil, fmt.Errorf("describe snapshots: %w", err)
	}

	now := time.Now().UTC()
	var results []DeadResource
	for _, snap := range snaps.Snapshots {
		if r, ok := considerSnapshot(snap, amiSnapshots, now, t, cfg.Region); ok {
			results = append(results, r)
		}
	}
	return results, nil
}

func (s *SnapshotScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

// considerSnapshot decides whether a single snapshot should be flagged.
// Snapshots referenced by any AMI are always skipped. A threshold of 0
// means "flag any snapshot with no AMI reference regardless of age".
func considerSnapshot(snap ec2types.Snapshot, amiSnapshots map[string]struct{}, now time.Time, t config.Thresholds, region string) (DeadResource, bool) {
	id := aws.ToString(snap.SnapshotId)
	if _, used := amiSnapshots[id]; used {
		return DeadResource{}, false
	}
	if snap.StartTime == nil {
		return DeadResource{}, false
	}
	age := now.Sub(*snap.StartTime)
	if t.SnapshotAgeDays > 0 {
		minAge := time.Duration(t.SnapshotAgeDays) * 24 * time.Hour
		if age < minAge {
			return DeadResource{}, false
		}
	}

	tags := make(map[string]string)
	for _, tag := range snap.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	var size int32
	if snap.VolumeSize != nil {
		size = *snap.VolumeSize
	}

	return DeadResource{
		Type:        "EBSSnapshot",
		ID:          id,
		Region:      region,
		Age:         age,
		MonthlyCost: float64(size) * 0.05,
		Reason:      snapshotReason(t.SnapshotAgeDays),
		Tags:        tags,
	}, true
}

func snapshotReason(days int) string {
	if days <= 0 {
		return "Snapshot with no AMI reference"
	}
	return fmt.Sprintf("Snapshot older than %d days with no AMI reference", days)
}
