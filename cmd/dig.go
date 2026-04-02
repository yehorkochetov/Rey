package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var digCmd = &cobra.Command{
	Use:   "dig",
	Short: "Scan AWS resources for unused or idle items",
	Long:  "Dig scans your AWS account for resources that may be unused, idle, or forgotten.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("scanning...")
	},
}

func init() {
	digCmd.Flags().Int("min-age", 7, "Minimum age in days to consider a resource idle")
	digCmd.Flags().String("export", "", "Export results to file (e.g. report.csv)")

	rootCmd.AddCommand(digCmd)
}
