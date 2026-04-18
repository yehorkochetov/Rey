package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

const rdsSnapshotMaxAge = 90 * 24 * time.Hour

type RDSSnapshotScanner struct{}

func (s *RDSSnapshotScanner) Name() string {
	return "rds-snapshot"
}

func (s *RDSSnapshotScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	client := rds.NewFromConfig(cfg)

	now := time.Now().UTC()
	var results []DeadResource

	paginator := rds.NewDescribeDBSnapshotsPaginator(client, &rds.DescribeDBSnapshotsInput{
		SnapshotType: aws.String("manual"),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe db snapshots: %w", err)
		}

		for _, snap := range page.DBSnapshots {
			if snap.SnapshotCreateTime == nil {
				continue
			}
			age := now.Sub(*snap.SnapshotCreateTime)
			if age < rdsSnapshotMaxAge {
				continue
			}

			id := aws.ToString(snap.DBSnapshotIdentifier)
			tags := make(map[string]string)
			for _, t := range snap.TagList {
				tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
			}

			var storage int32
			if snap.AllocatedStorage != nil {
				storage = *snap.AllocatedStorage
			}

			results = append(results, DeadResource{
				Type:        "RDSSnapshot",
				ID:          id,
				Name:        id,
				Region:      cfg.Region,
				Age:         age,
				MonthlyCost: float64(storage) * 0.095,
				Reason:      "Manual snapshot older than 90 days",
				Tags:        tags,
			})
		}
	}

	return results, nil
}

func (s *RDSSnapshotScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}
