package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yehorkochetov/rey/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "rey",
	Short: "Rey — AWS resource explorer",
	Long:  "Rey is a CLI tool for exploring and managing AWS resources.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.Init)

	rootCmd.PersistentFlags().String("region", "us-east-1", "AWS region")
	rootCmd.PersistentFlags().String("profile", "", "AWS profile")
	rootCmd.PersistentFlags().String("output", "table", "Output format (table, json, csv)")

	config.BindFlags(rootCmd)
}
