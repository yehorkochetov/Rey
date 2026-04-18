package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/yehorkochetov/rey/internal/config"
)

type RDSScanner struct{}

func (r *RDSScanner) Name() string {
	return "rds-instance"
}

func (r *RDSScanner) Scan(ctx context.Context, cfg aws.Config, _ config.Thresholds) ([]DeadResource, error) {
	client := rds.NewFromConfig(cfg)

	var results []DeadResource
	paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe db instances: %w", err)
		}

		for _, db := range page.DBInstances {
			if aws.ToString(db.DBInstanceStatus) != "stopped" {
				continue
			}

			id := aws.ToString(db.DBInstanceIdentifier)
			tags := make(map[string]string)
			for _, t := range db.TagList {
				tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
			}

			var storage int32
			if db.AllocatedStorage != nil {
				storage = *db.AllocatedStorage
			}

			results = append(results, DeadResource{
				Type:        "RDSInstance",
				ID:          id,
				Name:        id,
				Region:      cfg.Region,
				MonthlyCost: float64(storage) * 0.115,
				Reason:      "Stopped, still charging for storage",
				Tags:        tags,
			})
		}
	}

	return results, nil
}

func (r *RDSScanner) EstimateCost(res DeadResource) float64 {
	return res.MonthlyCost
}
