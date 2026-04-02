package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/viper"
)

func NewSession(ctx context.Context) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	if region := viper.GetString("region"); region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	if profile := viper.GetString("profile"); profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
