package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const multipartMaxAge = 7 * 24 * time.Hour

type S3MultipartScanner struct{}

func (s *S3MultipartScanner) Name() string {
	return "s3-multipart"
}

func (s *S3MultipartScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	client := s3.NewFromConfig(cfg)

	buckets, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	now := time.Now().UTC()
	var results []DeadResource

	for _, b := range buckets.Buckets {
		name := aws.ToString(b.Name)

		region, err := bucketRegion(ctx, client, name)
		if err != nil {
			return nil, err
		}
		if region != cfg.Region {
			continue
		}

		paginator := s3.NewListMultipartUploadsPaginator(client, &s3.ListMultipartUploadsInput{
			Bucket: aws.String(name),
		})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("list multipart uploads %s: %w", name, err)
			}

			for _, up := range page.Uploads {
				if up.Initiated == nil {
					continue
				}
				age := now.Sub(*up.Initiated)
				if age < multipartMaxAge {
					continue
				}

				key := aws.ToString(up.Key)
				uploadID := aws.ToString(up.UploadId)

				sizeBytes, err := multipartSize(ctx, client, name, key, uploadID)
				if err != nil {
					return nil, err
				}
				sizeGB := float64(sizeBytes) / (1024 * 1024 * 1024)

				results = append(results, DeadResource{
					Type:        "S3MultipartUpload",
					ID:          fmt.Sprintf("%s/%s", name, key),
					Name:        name,
					Region:      cfg.Region,
					Age:         age,
					MonthlyCost: sizeGB * 0.023,
					Reason:      "Incomplete multipart upload older than 7 days",
					Tags:        map[string]string{},
				})
			}
		}
	}

	return results, nil
}

func (s *S3MultipartScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func bucketRegion(ctx context.Context, client *s3.Client, bucket string) (string, error) {
	out, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return "", fmt.Errorf("get bucket location %s: %w", bucket, err)
	}
	region := string(out.LocationConstraint)
	if region == "" {
		return "us-east-1", nil
	}
	return region, nil
}

func multipartSize(ctx context.Context, client *s3.Client, bucket, key, uploadID string) (int64, error) {
	var total int64
	paginator := s3.NewListPartsPaginator(client, &s3.ListPartsInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, fmt.Errorf("list parts %s/%s: %w", bucket, key, err)
		}
		for _, p := range page.Parts {
			if p.Size != nil {
				total += *p.Size
			}
		}
	}
	return total, nil
}
