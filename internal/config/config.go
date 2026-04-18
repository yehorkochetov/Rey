package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Init() {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".rey")

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configDir)

	viper.SetEnvPrefix("REY")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}

func BindFlags(cmd *cobra.Command) {
	viper.BindPFlag("region", cmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("profile", cmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("output", cmd.PersistentFlags().Lookup("output"))
}

// LoadThresholds returns thresholds from the [thresholds] section of
// config.toml. Any key absent from the config falls back to its default —
// a present-but-zero value is preserved as an explicit zero.
func LoadThresholds() Thresholds {
	t := DefaultThresholds()
	apply := func(key string, dst *int) {
		if viper.IsSet("thresholds." + key) {
			*dst = viper.GetInt("thresholds." + key)
		}
	}
	apply("ec2_stopped_days", &t.EC2StoppedDays)
	apply("ebs_unattached_days", &t.EBSUnattachedDays)
	apply("snapshot_age_days", &t.SnapshotAgeDays)
	apply("dynamodb_idle_days", &t.DynamoDBIdleDays)
	apply("elasticache_idle_days", &t.ElastiCacheIdleDays)
	apply("nat_idle_days", &t.NATIdleDays)
	apply("s3_multipart_days", &t.S3MultipartDays)
	apply("s3_bucket_empty_days", &t.S3BucketEmptyDays)
	apply("ecr_image_age_days", &t.ECRImageAgeDays)
	apply("efs_idle_days", &t.EFSIdleDays)
	apply("cloudwatch_idle_days", &t.CloudWatchIdleDays)
	return t
}
