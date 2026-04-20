package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	reyaws "github.com/yehorkochetov/rey/internal/aws"
	"github.com/yehorkochetov/rey/internal/config"
	"github.com/yehorkochetov/rey/internal/output"
	"github.com/yehorkochetov/rey/internal/scanner"
)

var digCmd = &cobra.Command{
	Use:   "dig",
	Short: "Scan AWS resources for unused or idle items",
	Long:  "Dig scans your AWS account for resources that may be unused, idle, or forgotten.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := reyaws.NewSession(cmd.Context())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		reg := &scanner.Registry{}
		for _, s := range []scanner.Scanner{
			&scanner.EIPScanner{},
			&scanner.EC2Scanner{},
			&scanner.EBSScanner{},
			&scanner.SnapshotScanner{},
			&scanner.SecurityGroupScanner{},
			&scanner.ENIScanner{},
			&scanner.IGWScanner{},
			&scanner.NATGatewayScanner{},
			&scanner.VPCEndpointScanner{},
			&scanner.RDSScanner{},
			&scanner.RDSSnapshotScanner{},
			&scanner.ElastiCacheScanner{},
			&scanner.DynamoDBScanner{},
			&scanner.S3MultipartScanner{},
			&scanner.S3BucketScanner{},
			&scanner.ECRScanner{},
			&scanner.EFSScanner{},
		} {
			reg.Register(s)
		}

		thresholds := resolveThresholds(cmd)

		results, err := reg.RunAll(cmd.Context(), cfg, thresholds)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		output.RenderGraveyard(results)
	},
}

// resolveThresholds merges the threshold sources by priority:
// defaults -> config.toml -> CLI flags. A flag value of -1 means
// the user did not set it; 0 is a deliberate "no age check".
func resolveThresholds(cmd *cobra.Command) config.Thresholds {
	names := []string{
		"ec2-stopped-days",
		"ebs-unattached-days",
		"snapshot-age-days",
		"dynamodb-idle-days",
		"elasticache-idle-days",
		"nat-idle-days",
		"s3-multipart-days",
		"s3-bucket-empty-days",
		"ecr-image-age-days",
		"efs-idle-days",
	}
	flags := make(map[string]int, len(names))
	for _, n := range names {
		if v, err := cmd.Flags().GetInt(n); err == nil {
			flags[n] = v
		}
	}
	return config.ResolveThresholds(config.LoadThresholds(), flags)
}

func init() {
	digCmd.Flags().String("export", "", "Export results to file (e.g. report.csv)")

	digCmd.Flags().Int("ec2-stopped-days", -1, "Flag stopped EC2 instances older than N days (0 = any age)")
	digCmd.Flags().Int("ebs-unattached-days", -1, "Flag unattached EBS volumes older than N days (0 = any age)")
	digCmd.Flags().Int("snapshot-age-days", -1, "Flag EBS snapshots older than N days (0 = any age)")
	digCmd.Flags().Int("dynamodb-idle-days", -1, "Flag DynamoDB tables idle for N days (0 = any)")
	digCmd.Flags().Int("elasticache-idle-days", -1, "Flag ElastiCache clusters idle for N days (0 = any)")
	digCmd.Flags().Int("nat-idle-days", -1, "Flag NAT gateways idle for N days (0 = any)")
	digCmd.Flags().Int("s3-multipart-days", -1, "Flag incomplete multipart uploads older than N days (0 = any age)")
	digCmd.Flags().Int("s3-bucket-empty-days", -1, "Flag empty S3 buckets older than N days (0 = any age)")
	digCmd.Flags().Int("ecr-image-age-days", -1, "Flag untagged ECR images older than N days (0 = any age)")
	digCmd.Flags().Int("efs-idle-days", -1, "Flag EFS file systems idle for N days (0 = any)")

	rootCmd.AddCommand(digCmd)
}
