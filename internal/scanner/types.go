package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/yehorkochetov/rey/internal/config"
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
	Scan(ctx context.Context, cfg aws.Config, t config.Thresholds) ([]DeadResource, error)
	EstimateCost(r DeadResource) float64
}

// idleReason builds the human-readable reason for an idle resource.
// A non-positive day count means the threshold is disabled, so the
// reason drops the day suffix.
func idleReason(prefix string, days int) string {
	if days <= 0 {
		return prefix
	}
	return fmt.Sprintf("%s in %d days", prefix, days)
}
