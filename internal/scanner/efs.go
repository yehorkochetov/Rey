package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"

	"github.com/yehorkochetov/rey/internal/config"
)

const efsIdleWindow = 7 * 24 * time.Hour

type EFSScanner struct{}

func (e *EFSScanner) Name() string {
	return "efs"
}

func (e *EFSScanner) Scan(ctx context.Context, cfg aws.Config, _ config.Thresholds) ([]DeadResource, error) {
	efsClient := efs.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	now := time.Now().UTC()
	start := now.Add(-efsIdleWindow)

	var results []DeadResource
	paginator := efs.NewDescribeFileSystemsPaginator(efsClient, &efs.DescribeFileSystemsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe file systems: %w", err)
		}

		for _, fs := range page.FileSystems {
			id := aws.ToString(fs.FileSystemId)

			io, err := efsMeteredIO(ctx, cwClient, id, start, now)
			if err != nil {
				return nil, err
			}
			if io > 0 {
				continue
			}

			tags := make(map[string]string)
			var name string
			for _, t := range fs.Tags {
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

			var sizeGB float64
			if fs.SizeInBytes != nil {
				sizeGB = float64(fs.SizeInBytes.Value) / 1073741824
			}

			results = append(results, DeadResource{
				Type:        "EFSFileSystem",
				ID:          id,
				Name:        name,
				Region:      cfg.Region,
				MonthlyCost: sizeGB * 0.30,
				Reason:      "No IO activity in 7 days",
				Tags:        tags,
			})
		}
	}

	return results, nil
}

func (e *EFSScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func efsMeteredIO(ctx context.Context, client *cloudwatch.Client, fsID string, start, end time.Time) (float64, error) {
	stats, err := client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/EFS"),
		MetricName: aws.String("MeteredIOBytes"),
		Dimensions: []cwtypes.Dimension{
			{Name: aws.String("FileSystemId"), Value: aws.String(fsID)},
		},
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		Period:     aws.Int32(int32(efsIdleWindow.Seconds())),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
	})
	if err != nil {
		return 0, fmt.Errorf("efs metrics %s: %w", fsID, err)
	}

	var total float64
	for _, dp := range stats.Datapoints {
		if dp.Sum != nil {
			total += *dp.Sum
		}
	}
	return total, nil
}
