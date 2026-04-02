package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/viper"
)

// NewSession loads AWS config with the following priority:
// flag > env var > ~/.rey/config.toml > default credential chain
func NewSession(ctx context.Context) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithSharedConfigFiles([]string{
			config.DefaultSharedConfigFilename(),
		}),
		config.WithSharedCredentialsFiles([]string{
			config.DefaultSharedCredentialsFilename(),
		}),
	}

	// viper resolves flag > env > config.toml > default automatically
	if region := viper.GetString("region"); region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	if profile := viper.GetString("profile"); profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
