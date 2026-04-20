package config

import "testing"

func TestDefaultThresholds_Values(t *testing.T) {
	got := DefaultThresholds()
	want := Thresholds{
		EC2StoppedDays:      7,
		EBSUnattachedDays:   0,
		SnapshotAgeDays:     90,
		DynamoDBIdleDays:    14,
		ElastiCacheIdleDays: 7,
		NATIdleDays:         7,
		S3MultipartDays:     7,
		S3BucketEmptyDays:   30,
		ECRImageAgeDays:     180,
		EFSIdleDays:         7,
		CloudWatchIdleDays:  30,
	}
	if got != want {
		t.Errorf("DefaultThresholds() = %+v, want %+v", got, want)
	}
}

// The EBSUnattachedDays default is intentionally 0 ("flag every unattached
// volume regardless of age"). This test asserts every *other* field is
// non-zero so a future refactor can't silently collapse a meaningful default
// to the struct zero value.
func TestDefaultThresholds_NonZeroFields(t *testing.T) {
	d := DefaultThresholds()
	cases := []struct {
		name string
		v    int
	}{
		{"EC2StoppedDays", d.EC2StoppedDays},
		{"SnapshotAgeDays", d.SnapshotAgeDays},
		{"DynamoDBIdleDays", d.DynamoDBIdleDays},
		{"ElastiCacheIdleDays", d.ElastiCacheIdleDays},
		{"NATIdleDays", d.NATIdleDays},
		{"S3MultipartDays", d.S3MultipartDays},
		{"S3BucketEmptyDays", d.S3BucketEmptyDays},
		{"ECRImageAgeDays", d.ECRImageAgeDays},
		{"EFSIdleDays", d.EFSIdleDays},
		{"CloudWatchIdleDays", d.CloudWatchIdleDays},
	}
	for _, c := range cases {
		if c.v == 0 {
			t.Errorf("%s default is zero; expected a positive number of days", c.name)
		}
	}
}

func TestResolveThresholds_NoFlagsReturnsBase(t *testing.T) {
	base := DefaultThresholds()
	got := ResolveThresholds(base, nil)
	if got != base {
		t.Errorf("ResolveThresholds(base, nil) = %+v, want %+v", got, base)
	}
}

func TestResolveThresholds_ConfigValueFlowsThrough(t *testing.T) {
	// A config-loaded base with a non-default value should survive when no
	// CLI flag is provided for that key.
	base := DefaultThresholds()
	base.EC2StoppedDays = 42
	got := ResolveThresholds(base, map[string]int{})
	if got.EC2StoppedDays != 42 {
		t.Errorf("config value lost: got %d, want 42", got.EC2StoppedDays)
	}
}

func TestResolveThresholds_FlagMinusOneDoesNotOverride(t *testing.T) {
	base := DefaultThresholds()
	base.EC2StoppedDays = 42
	got := ResolveThresholds(base, map[string]int{"ec2-stopped-days": -1})
	if got.EC2StoppedDays != 42 {
		t.Errorf("-1 flag wrongly overrode config: got %d, want 42", got.EC2StoppedDays)
	}
}

func TestResolveThresholds_FlagZeroOverridesConfig(t *testing.T) {
	// 0 is a deliberate "no age check" — it must replace the config value.
	base := DefaultThresholds()
	base.EC2StoppedDays = 42
	got := ResolveThresholds(base, map[string]int{"ec2-stopped-days": 0})
	if got.EC2StoppedDays != 0 {
		t.Errorf("0 flag did not override: got %d, want 0", got.EC2StoppedDays)
	}
}

func TestResolveThresholds_LargeFlagOverridesConfig(t *testing.T) {
	base := DefaultThresholds()
	base.EC2StoppedDays = 42
	got := ResolveThresholds(base, map[string]int{"ec2-stopped-days": 99999})
	if got.EC2StoppedDays != 99999 {
		t.Errorf("large flag did not override: got %d, want 99999", got.EC2StoppedDays)
	}
}

func TestResolveThresholds_ZeroStaysZero(t *testing.T) {
	// Caller passed 0 (intentional no-age) — ResolveThresholds must not
	// treat it as "missing" and fall back to a default.
	base := Thresholds{EC2StoppedDays: 0}
	got := ResolveThresholds(base, nil)
	if got.EC2StoppedDays != 0 {
		t.Errorf("zero replaced: got %d, want 0", got.EC2StoppedDays)
	}
}

func TestResolveThresholds_AllFlagsRouteToCorrectField(t *testing.T) {
	// Guards against silent bugs like mapping "nat-idle-days" to EFSIdleDays.
	flags := map[string]int{
		"ec2-stopped-days":      1,
		"ebs-unattached-days":   2,
		"snapshot-age-days":     3,
		"dynamodb-idle-days":    4,
		"elasticache-idle-days": 5,
		"nat-idle-days":         6,
		"s3-multipart-days":     7,
		"s3-bucket-empty-days":  8,
		"ecr-image-age-days":    9,
		"efs-idle-days":         10,
		"cloudwatch-idle-days":  11,
	}
	got := ResolveThresholds(Thresholds{}, flags)
	want := Thresholds{
		EC2StoppedDays:      1,
		EBSUnattachedDays:   2,
		SnapshotAgeDays:     3,
		DynamoDBIdleDays:    4,
		ElastiCacheIdleDays: 5,
		NATIdleDays:         6,
		S3MultipartDays:     7,
		S3BucketEmptyDays:   8,
		ECRImageAgeDays:     9,
		EFSIdleDays:         10,
		CloudWatchIdleDays:  11,
	}
	if got != want {
		t.Errorf("ResolveThresholds per-field routing wrong:\n got=%+v\nwant=%+v", got, want)
	}
}
