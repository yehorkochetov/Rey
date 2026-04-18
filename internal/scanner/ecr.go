package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"

	"github.com/yehorkochetov/rey/internal/config"
)

const ecrImageMaxAge = 180 * 24 * time.Hour

type ECRScanner struct{}

func (e *ECRScanner) Name() string {
	return "ecr-image"
}

func (e *ECRScanner) Scan(ctx context.Context, cfg aws.Config, _ config.Thresholds) ([]DeadResource, error) {
	client := ecr.NewFromConfig(cfg)

	now := time.Now().UTC()
	var results []DeadResource

	repos := ecr.NewDescribeRepositoriesPaginator(client, &ecr.DescribeRepositoriesInput{})
	for repos.HasMorePages() {
		repoPage, err := repos.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe repositories: %w", err)
		}

		for _, r := range repoPage.Repositories {
			repoName := aws.ToString(r.RepositoryName)

			images := ecr.NewDescribeImagesPaginator(client, &ecr.DescribeImagesInput{
				RepositoryName: aws.String(repoName),
			})
			for images.HasMorePages() {
				imgPage, err := images.NextPage(ctx)
				if err != nil {
					return nil, fmt.Errorf("describe images %s: %w", repoName, err)
				}

				for _, img := range imgPage.ImageDetails {
					if img.ImagePushedAt == nil {
						continue
					}
					age := now.Sub(*img.ImagePushedAt)
					if age < ecrImageMaxAge {
						continue
					}
					if hasLatestTag(img.ImageTags) {
						continue
					}

					digest := aws.ToString(img.ImageDigest)
					short := digest
					if len(short) > 12 {
						short = short[:12]
					}

					var sizeGB float64
					if img.ImageSizeInBytes != nil {
						sizeGB = float64(*img.ImageSizeInBytes) / 1073741824
					}

					results = append(results, DeadResource{
						Type:        "ECRImage",
						ID:          fmt.Sprintf("%s:%s", repoName, short),
						Name:        repoName,
						Region:      cfg.Region,
						Age:         age,
						MonthlyCost: sizeGB * 0.10,
						Reason:      "Untagged ECR image older than 180 days",
						Tags:        map[string]string{},
					})
				}
			}
		}
	}

	return results, nil
}

func (e *ECRScanner) EstimateCost(r DeadResource) float64 {
	return r.MonthlyCost
}

func hasLatestTag(tags []string) bool {
	for _, t := range tags {
		if t == "latest" {
			return true
		}
	}
	return false
}
