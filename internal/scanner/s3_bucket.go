package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/yehorkochetov/rey/internal/config"
)

type S3BucketScanner struct{}

func (s *S3BucketScanner) Name() string {
	return "s3-bucket"
}

func (s *S3BucketScanner) Scan(ctx context.Context, cfg aws.Config, t config.Thresholds) ([]DeadResource, error) {
	minAge := time.Duration(t.S3BucketEmptyDays) * 24 * time.Hour
	client := s3.NewFromConfig(cfg)

	buckets, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	now := time.Now().UTC()
	var results []DeadResource

	for _, b := range buckets.Buckets {
		name := aws.ToString(b.Name)
		var age time.Duration
		if b.CreationDate != nil {
			age = now.Sub(*b.CreationDate)
		}
		if t.S3BucketEmptyDays > 0 && age < minAge {
			continue
		}

		region, err := bucketRegion(ctx, client, name)
		if err != nil {
			return nil, err
		}
		if region != cfg.Region {
			continue
		}

		objs, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:  aws.String(name),
			MaxKeys: aws.Int32(1),
		})
		if err != nil {
			return nil, fmt.Errorf("list objects %s: %w", name, err)
		}
		if aws.ToInt32(objs.KeyCount) > 0 {
			continue
		}

		results = append(results, DeadResource{
			Type:        "S3Bucket",
			ID:          name,
			Name:        name,
			Region:      cfg.Region,
			Age:         age,
			MonthlyCost: 0,
			Reason:      emptyBucketReason(t.S3BucketEmptyDays),
			Tags:        map[string]string{},
		})
	}

	return results, nil
}

func (s *S3BucketScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func emptyBucketReason(days int) string {
	if days <= 0 {
		return "Empty bucket"
	}
	return fmt.Sprintf("Empty bucket older than %d days", days)
}
