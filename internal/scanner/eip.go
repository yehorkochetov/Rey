package scanner

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type EIPScanner struct{}

func (e *EIPScanner) Name() string {
	return "elastic-ip"
}

func (e *EIPScanner) Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error) {
	return nil, nil
}

func (e *EIPScanner) EstimateCost(r DeadResource) float64 {
	return 0
}
