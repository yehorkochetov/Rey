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
