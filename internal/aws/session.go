package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, wrapConfigError(err)
	}

	stsClient := sts.NewFromConfig(cfg)
	if _, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}); err != nil {
		return cfg, wrapSTSError(err)
	}

	return cfg, nil
}

func wrapConfigError(err error) error {
	msg := err.Error()

	if strings.Contains(msg, "could not find profile") || strings.Contains(msg, "failed to get shared config profile") {
		profile := viper.GetString("profile")
		return fmt.Errorf("profile '%s' not found in ~/.aws/config", profile)
	}

	return err
}

func wrapSTSError(err error) error {
	msg := err.Error()

	if strings.Contains(msg, "no EC2 IMDS role found") ||
		strings.Contains(msg, "failed to refresh cached credentials") ||
		strings.Contains(msg, "NoCredentialProviders") {
		return fmt.Errorf("AWS credentials not found. Run 'aws configure'")
	}

	if strings.Contains(msg, "ExpiredToken") || strings.Contains(msg, "expired") {
		return fmt.Errorf("AWS session token expired, please refresh")
	}

	if strings.Contains(msg, "could not find profile") || strings.Contains(msg, "SharedConfigProfileNotExist") {
		profile := viper.GetString("profile")
		return fmt.Errorf("profile '%s' not found in ~/.aws/config", profile)
	}

	return err
}
