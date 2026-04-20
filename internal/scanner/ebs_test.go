package scanner

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

func TestConsiderEBSVolume_AgeThreshold(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name          string
		createdDaysBack int
		threshold     int
		wantFlagged   bool
	}{
		{"newer than threshold is skipped", 3, 7, false},
		{"exactly at threshold is flagged", 7, 7, true},
		{"older than threshold is flagged", 30, 7, true},
		{"threshold 0 flags any age", 1, 0, true},
		{"threshold 0 flags zero age", 0, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			created := now.Add(-time.Duration(c.createdDaysBack) * 24 * time.Hour)
			v := ec2types.Volume{
				VolumeId:   aws.String("vol-test"),
				CreateTime: &created,
				Size:       aws.Int32(10),
			}
			_, ok := considerEBSVolume(v, now, config.Thresholds{EBSUnattachedDays: c.threshold}, "us-east-1")
			if ok != c.wantFlagged {
				t.Errorf("flagged=%v, want %v", ok, c.wantFlagged)
			}
		})
	}
}

func TestConsiderEBSVolume_CostFromSize(t *testing.T) {
	now := time.Now().UTC()
	created := now.Add(-30 * 24 * time.Hour)
	cases := []struct {
		name    string
		sizeGB  int32
		wantUSD float64
	}{
		{"100GB at $0.10/GB", 100, 10.00},
		{"1GB", 1, 0.10},
		{"500GB", 500, 50.00},
		{"0GB (nil-like)", 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := ec2types.Volume{
				VolumeId:   aws.String("vol-cost"),
				CreateTime: &created,
				Size:       aws.Int32(c.sizeGB),
			}
			r, ok := considerEBSVolume(v, now, config.Thresholds{EBSUnattachedDays: 0}, "us-east-1")
			if !ok {
				t.Fatal("expected flagged")
			}
			if r.MonthlyCost != c.wantUSD {
				t.Errorf("cost=%.4f, want %.4f", r.MonthlyCost, c.wantUSD)
			}
		})
	}
}

func TestConsiderEBSVolume_NilSizeIsSafe(t *testing.T) {
	now := time.Now().UTC()
	created := now.Add(-30 * 24 * time.Hour)
	v := ec2types.Volume{
		VolumeId:   aws.String("vol-nil-size"),
		CreateTime: &created,
		Size:       nil,
	}
	r, ok := considerEBSVolume(v, now, config.Thresholds{EBSUnattachedDays: 0}, "us-east-1")
	if !ok {
		t.Fatal("expected flagged")
	}
	if r.MonthlyCost != 0 {
		t.Errorf("cost with nil size = %v, want 0", r.MonthlyCost)
	}
}
