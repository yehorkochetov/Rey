package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	reyaws "github.com/yehorkochetov/rey/internal/aws"
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
		results, err := reg.RunAll(cmd.Context(), cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d resources\n", len(results))
	},
}

func init() {
	digCmd.Flags().Int("min-age", 7, "Minimum age in days to consider a resource idle")
	digCmd.Flags().String("export", "", "Export results to file (e.g. report.csv)")

	rootCmd.AddCommand(digCmd)
}
