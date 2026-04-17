package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

const snapshotMaxAge = 90 * 24 * time.Hour

type SnapshotScanner struct{}

func (s *SnapshotScanner) Name() string {
	return "ebs-snapshot"
}

func (s *SnapshotScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
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

	var results []DeadResource
	now := time.Now().UTC()
	for _, snap := range snaps.Snapshots {
		id := aws.ToString(snap.SnapshotId)
		if _, used := amiSnapshots[id]; used {
			continue
		}
		if snap.StartTime == nil {
			continue
		}
		age := now.Sub(*snap.StartTime)
		if age < snapshotMaxAge {
			continue
		}

		tags := make(map[string]string)
		for _, t := range snap.Tags {
			tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
		}
		var size int32
		if snap.VolumeSize != nil {
			size = *snap.VolumeSize
		}

		results = append(results, DeadResource{
			Type:        "EBSSnapshot",
			ID:          id,
			Region:      cfg.Region,
			Age:         age,
			MonthlyCost: float64(size) * 0.05,
			Reason:      "Snapshot older than 90 days with no AMI reference",
			Tags:        tags,
		})
	}

	return results, nil
}

func (s *SnapshotScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}
