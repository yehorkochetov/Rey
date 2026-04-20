package config

// Thresholds holds the per-scanner age/idle thresholds in days.
// Zero is a valid value meaning "flag every matching resource regardless
// of age" — never treat 0 as a missing value.
type Thresholds struct {
	EC2StoppedDays      int
	EBSUnattachedDays   int
	SnapshotAgeDays     int
	DynamoDBIdleDays    int
	ElastiCacheIdleDays int
	NATIdleDays         int
	S3MultipartDays     int
	S3BucketEmptyDays   int
	ECRImageAgeDays     int
	EFSIdleDays         int
	CloudWatchIdleDays  int
}

func DefaultThresholds() Thresholds {
	return Thresholds{
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
}

// ResolveThresholds layers CLI flag values on top of the config-loaded base.
// A flag value of -1 means the user did not set it, so the base wins; any
// other value — including 0 — replaces the base.
func ResolveThresholds(base Thresholds, flags map[string]int) Thresholds {
	t := base
	apply := func(name string, dst *int) {
		v, ok := flags[name]
		if !ok || v == -1 {
			return
		}
		*dst = v
	}
	apply("ec2-stopped-days", &t.EC2StoppedDays)
	apply("ebs-unattached-days", &t.EBSUnattachedDays)
	apply("snapshot-age-days", &t.SnapshotAgeDays)
	apply("dynamodb-idle-days", &t.DynamoDBIdleDays)
	apply("elasticache-idle-days", &t.ElastiCacheIdleDays)
	apply("nat-idle-days", &t.NATIdleDays)
	apply("s3-multipart-days", &t.S3MultipartDays)
	apply("s3-bucket-empty-days", &t.S3BucketEmptyDays)
	apply("ecr-image-age-days", &t.ECRImageAgeDays)
	apply("efs-idle-days", &t.EFSIdleDays)
	apply("cloudwatch-idle-days", &t.CloudWatchIdleDays)
	return t
}
