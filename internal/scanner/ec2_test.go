package scanner

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/yehorkochetov/rey/internal/config"
)

// stopReason builds the EC2 StateTransitionReason string format that
// parseStopTime expects: "User initiated (YYYY-MM-DD HH:MM:SS GMT)".
func stopReason(stoppedAt time.Time) string {
	return fmt.Sprintf("User initiated (%s GMT)", stoppedAt.Format("2006-01-02 15:04:05"))
}

func TestConsiderEC2Instance_RunningNeverFlagged(t *testing.T) {
	// Even if the API filter slipped a running instance through, the pure
	// filter must reject it.
	inst := ec2types.Instance{
		InstanceId:            aws.String("i-running"),
		State:                 &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning},
		StateTransitionReason: aws.String(stopReason(time.Now().UTC().Add(-30 * 24 * time.Hour))),
	}
	_, ok := considerEC2Instance(inst, time.Now().UTC(), config.Thresholds{EC2StoppedDays: 7}, "us-east-1")
	if ok {
		t.Error("running instance must never be flagged")
	}
}

func TestConsiderEC2Instance_StoppedAgeThreshold(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name         string
		stoppedDays  int
		threshold    int
		wantFlagged  bool
	}{
		{"stopped 2 days, threshold 7: skipped", 2, 7, false},
		{"stopped 7 days, threshold 7: flagged", 7, 7, true},
		{"stopped 30 days, threshold 7: flagged", 30, 7, true},
		{"threshold 0 flags regardless of age", 1, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stoppedAt := now.Add(-time.Duration(c.stoppedDays) * 24 * time.Hour)
			inst := ec2types.Instance{
				InstanceId:            aws.String("i-test"),
				State:                 &ec2types.InstanceState{Name: ec2types.InstanceStateNameStopped},
				StateTransitionReason: aws.String(stopReason(stoppedAt)),
			}
			_, ok := considerEC2Instance(inst, now, config.Thresholds{EC2StoppedDays: c.threshold}, "us-east-1")
			if ok != c.wantFlagged {
				t.Errorf("flagged=%v, want %v", ok, c.wantFlagged)
			}
		})
	}
}

func TestConsiderEC2Instance_UnparseableStopReasonSkipped(t *testing.T) {
	// Instances stopped by the scheduler, or older AWS API formats, may
	// lack a parseable timestamp. We can't compute age, so we skip rather
	// than treat as age=0 and flag everything.
	inst := ec2types.Instance{
		InstanceId:            aws.String("i-weird"),
		State:                 &ec2types.InstanceState{Name: ec2types.InstanceStateNameStopped},
		StateTransitionReason: aws.String("Server.ScheduledStop"),
	}
	_, ok := considerEC2Instance(inst, time.Now().UTC(), config.Thresholds{EC2StoppedDays: 0}, "us-east-1")
	if ok {
		t.Error("instance with unparseable StateTransitionReason should be skipped")
	}
}

func TestConsiderEC2Instance_NilStateSkipped(t *testing.T) {
	inst := ec2types.Instance{
		InstanceId: aws.String("i-no-state"),
		State:      nil,
	}
	_, ok := considerEC2Instance(inst, time.Now().UTC(), config.Thresholds{EC2StoppedDays: 0}, "us-east-1")
	if ok {
		t.Error("instance with nil State should be skipped")
	}
}
