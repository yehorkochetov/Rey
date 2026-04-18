package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/yehorkochetov/rey/internal/config"
)

const elastiCacheIdleWindow = 7 * 24 * time.Hour

type ElastiCacheScanner struct{}

func (e *ElastiCacheScanner) Name() string {
	return "elasticache"
}

func (e *ElastiCacheScanner) Scan(ctx context.Context, cfg aws.Config, _ config.Thresholds) ([]DeadResource, error) {
	ecClient := elasticache.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	now := time.Now().UTC()
	start := now.Add(-elastiCacheIdleWindow)

	var results []DeadResource
	paginator := elasticache.NewDescribeCacheClustersPaginator(ecClient, &elasticache.DescribeCacheClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe cache clusters: %w", err)
		}

		for _, c := range page.CacheClusters {
			id := aws.ToString(c.CacheClusterId)

			avg, err := elastiCacheConnections(ctx, cwClient, id, start, now)
			if err != nil {
				return nil, err
			}
			if avg > 0 {
				continue
			}

			results = append(results, DeadResource{
				Type:        "ElastiCacheCluster",
				ID:          id,
				Name:        id,
				Region:      cfg.Region,
				MonthlyCost: 0,
				Reason:      "No connections in 7 days",
				Tags:        map[string]string{},
			})
		}
	}

	return results, nil
}

func (e *ElastiCacheScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func elastiCacheConnections(ctx context.Context, client *cloudwatch.Client, clusterID string, start, end time.Time) (float64, error) {
	stats, err := client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/ElastiCache"),
		MetricName: aws.String("CurrConnections"),
		Dimensions: []cwtypes.Dimension{
			{Name: aws.String("CacheClusterId"), Value: aws.String(clusterID)},
		},
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		Period:     aws.Int32(int32(elastiCacheIdleWindow.Seconds())),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticAverage},
	})
	if err != nil {
		return 0, fmt.Errorf("elasticache metrics %s: %w", clusterID, err)
	}

	var max float64
	for _, dp := range stats.Datapoints {
		if dp.Average != nil && *dp.Average > max {
			max = *dp.Average
		}
	}
	return max, nil
}
