package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	reyaws "github.com/yehorkochetov/rey/internal/aws"
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

		minAgeDays, _ := cmd.Flags().GetInt("min-age")
		minAge := time.Duration(minAgeDays) * 24 * time.Hour

		reg := &scanner.Registry{}
		reg.Register(&scanner.EIPScanner{})
		reg.Register(&scanner.EC2Scanner{MinAge: minAge})
		reg.Register(&scanner.EBSScanner{MinAge: minAge})
		reg.Register(&scanner.SnapshotScanner{})

		results, err := reg.RunAll(cmd.Context(), cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		output.RenderGraveyard(results)
	},
}

func init() {
	digCmd.Flags().Int("min-age", 7, "Minimum age in days to consider a resource idle")
	digCmd.Flags().String("export", "", "Export results to file (e.g. report.csv)")

	rootCmd.AddCommand(digCmd)
}
