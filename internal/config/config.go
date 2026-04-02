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
