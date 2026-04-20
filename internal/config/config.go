package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ExampleConfigTOML is the documented schema for ~/.rey/config.toml.
// Users can drop this into the file to override per-scanner thresholds;
// any key omitted falls back to DefaultThresholds().
const ExampleConfigTOML = `# ~/.rey/config.toml — rey configuration

region  = "us-east-1"
profile = ""
output  = "table"

# Per-scanner thresholds in days. Zero means "flag every matching
# resource regardless of age" — never treat zero as missing.
[thresholds]
ec2_stopped_days      = 7
ebs_unattached_days   = 0
snapshot_age_days     = 90
dynamodb_idle_days    = 14
elasticache_idle_days = 7
nat_idle_days         = 7
s3_multipart_days     = 7
s3_bucket_empty_days  = 30
ecr_image_age_days    = 180
efs_idle_days         = 7
cloudwatch_idle_days  = 30
`

func Init() {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".rey")

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configDir)

	viper.SetEnvPrefix("REY")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()

	// Keys nested under [defaults] in config.toml act as fallbacks for
	// the top-level keys the CLI binds against. SetDefault sits below
	// flags and env, so an explicit --profile still wins.
	for _, k := range []string{"region", "profile", "output"} {
		if v := viper.GetString("defaults." + k); v != "" {
			viper.SetDefault(k, v)
		}
	}
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
