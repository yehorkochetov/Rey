package scanner

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestConsiderSecurityGroup_DefaultGroupSkipped(t *testing.T) {
	g := ec2types.SecurityGroup{
		GroupId:   aws.String("sg-default"),
		GroupName: aws.String("default"),
	}
	_, ok := considerSecurityGroup(g, nil, "us-east-1")
	if ok {
		t.Error(`security group named "default" must never be flagged`)
	}
}

func TestConsiderSecurityGroup_AttachedToENISkipped(t *testing.T) {
	g := ec2types.SecurityGroup{
		GroupId:   aws.String("sg-123"),
		GroupName: aws.String("my-app"),
	}
	inUse := map[string]struct{}{"sg-123": {}}
	_, ok := considerSecurityGroup(g, inUse, "us-east-1")
	if ok {
		t.Error("group attached to at least one ENI must not be flagged")
	}
}

func TestConsiderSecurityGroup_UnattachedIsFlagged(t *testing.T) {
	g := ec2types.SecurityGroup{
		GroupId:   aws.String("sg-orphan"),
		GroupName: aws.String("my-unused"),
		Tags: []ec2types.Tag{
			{Key: aws.String("Name"), Value: aws.String("unused-sg")},
		},
	}
	r, ok := considerSecurityGroup(g, map[string]struct{}{"sg-other": {}}, "us-east-1")
	if !ok {
		t.Fatal("unattached non-default group should be flagged")
	}
	if r.Type != "SecurityGroup" {
		t.Errorf("Type = %q, want SecurityGroup", r.Type)
	}
	if r.ID != "sg-orphan" {
		t.Errorf("ID = %q, want sg-orphan", r.ID)
	}
	if r.MonthlyCost != 0 {
		t.Errorf("cost = %v, want 0 (security groups are free)", r.MonthlyCost)
	}
}

func TestConsiderSecurityGroup_NilInUseMap(t *testing.T) {
	// Defensive: if the scanner hasn't populated inUse (e.g. no ENIs exist)
	// passing nil must still work.
	g := ec2types.SecurityGroup{
		GroupId:   aws.String("sg-1"),
		GroupName: aws.String("foo"),
	}
	_, ok := considerSecurityGroup(g, nil, "us-east-1")
	if !ok {
		t.Error("with nil inUse map, non-default group should be flagged")
	}
}
