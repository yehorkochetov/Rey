// internal/scanner/types.go
package scanner

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// DeadResource represents a single wasted or idle AWS resource.
type DeadResource struct {
	ID          string
	Type        string
	Region      string
	Name        string
	Age         time.Duration
	MonthlyCost float64
	Reason      string
	Tags        map[string]string
}

// Scanner is implemented by every resource scanner.
// Each scanner is responsible for one resource type.
type Scanner interface {
	Name() string
	Scan(ctx context.Context, cfg aws.Config) ([]DeadResource, error)
	EstimateCost(r DeadResource) float64
}
