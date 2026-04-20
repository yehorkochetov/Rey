package scanner

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

func TestConsiderSnapshot_AgeThreshold(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name            string
		startedDaysBack int
		threshold       int
		wantFlagged     bool
	}{
		{"newer than 90d skipped", 30, 90, false},
		{"exactly 90d flagged", 90, 90, true},
		{"older than 90d flagged", 365, 90, true},
		{"threshold 0 flags any age", 1, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			start := now.Add(-time.Duration(c.startedDaysBack) * 24 * time.Hour)
			snap := ec2types.Snapshot{
				SnapshotId: aws.String("snap-test"),
				StartTime:  &start,
				VolumeSize: aws.Int32(10),
			}
			_, ok := considerSnapshot(snap, nil, now, config.Thresholds{SnapshotAgeDays: c.threshold}, "us-east-1")
			if ok != c.wantFlagged {
				t.Errorf("flagged=%v, want %v", ok, c.wantFlagged)
			}
		})
	}
}

func TestConsiderSnapshot_AMIReferenceNeverFlagged(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(-365 * 24 * time.Hour) // very old — would otherwise flag
	snap := ec2types.Snapshot{
		SnapshotId: aws.String("snap-used-by-ami"),
		StartTime:  &start,
		VolumeSize: aws.Int32(10),
	}
	amiSnapshots := map[string]struct{}{"snap-used-by-ami": {}}
	_, ok := considerSnapshot(snap, amiSnapshots, now, config.Thresholds{SnapshotAgeDays: 90}, "us-east-1")
	if ok {
		t.Error("snapshot referenced by AMI should never be flagged")
	}
}

func TestConsiderSnapshot_UnreferencedIsFlagged(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(-200 * 24 * time.Hour)
	snap := ec2types.Snapshot{
		SnapshotId: aws.String("snap-orphan"),
		StartTime:  &start,
		VolumeSize: aws.Int32(20),
	}
	// amiSnapshots contains a different id — this snapshot is unreferenced.
	amiSnapshots := map[string]struct{}{"snap-other": {}}
	r, ok := considerSnapshot(snap, amiSnapshots, now, config.Thresholds{SnapshotAgeDays: 90}, "us-east-1")
	if !ok {
		t.Fatal("unreferenced old snapshot should be flagged")
	}
	if r.Type != "EBSSnapshot" {
		t.Errorf("Type = %q, want EBSSnapshot", r.Type)
	}
}

func TestConsiderSnapshot_NilStartTimeSkipped(t *testing.T) {
	// An AWS snapshot with no StartTime is malformed — we skip it rather
	// than treat it as infinitely old.
	snap := ec2types.Snapshot{
		SnapshotId: aws.String("snap-no-time"),
		StartTime:  nil,
	}
	_, ok := considerSnapshot(snap, nil, time.Now().UTC(), config.Thresholds{SnapshotAgeDays: 0}, "us-east-1")
	if ok {
		t.Error("snapshot with nil StartTime should be skipped")
	}
}
