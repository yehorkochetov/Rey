package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const dynamoIdleWindow = 14 * 24 * time.Hour

type DynamoDBScanner struct{}

func (d *DynamoDBScanner) Name() string {
	return "dynamodb"
}

func (d *DynamoDBScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	ddbClient := dynamodb.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	now := time.Now().UTC()
	start := now.Add(-dynamoIdleWindow)

	var results []DeadResource
	paginator := dynamodb.NewListTablesPaginator(ddbClient, &dynamodb.ListTablesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list dynamodb tables: %w", err)
		}

		for _, name := range page.TableNames {
			reads, err := dynamoCapacity(ctx, cwClient, name, "ConsumedReadCapacityUnits", start, now)
			if err != nil {
				return nil, err
			}
			writes, err := dynamoCapacity(ctx, cwClient, name, "ConsumedWriteCapacityUnits", start, now)
			if err != nil {
				return nil, err
			}
			if reads > 0 || writes > 0 {
				continue
			}

			results = append(results, DeadResource{
				Type:        "DynamoDBTable",
				ID:          name,
				Name:        name,
				Region:      cfg.Region,
				MonthlyCost: 0,
				Reason:      "No reads or writes in 14 days",
				Tags:        map[string]string{},
			})
		}
	}

	return results, nil
}

func (d *DynamoDBScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func dynamoCapacity(ctx context.Context, client *cloudwatch.Client, table, metric string, start, end time.Time) (float64, error) {
	stats, err := client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/DynamoDB"),
		MetricName: aws.String(metric),
		Dimensions: []cwtypes.Dimension{
			{Name: aws.String("TableName"), Value: aws.String(table)},
		},
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		Period:     aws.Int32(int32(dynamoIdleWindow.Seconds())),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
	})
	if err != nil {
		return 0, fmt.Errorf("dynamodb metrics %s/%s: %w", table, metric, err)
	}

	var total float64
	for _, dp := range stats.Datapoints {
		if dp.Sum != nil {
			total += *dp.Sum
		}
	}
	return total, nil
}
