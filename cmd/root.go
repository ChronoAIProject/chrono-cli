package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ChronoAIProject/chrono-cli/pkg/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "chrono",
	Short: "Chrono CLI - Developer Platform command-line interface",
	Long: `Chrono CLI allows you to interact with the Developer Platform without accessing the dashboard.

Features:
- Authentication via Keycloak device flow
- MCP server configuration for AI editors
- Project detection and analysis`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set version from package variable
	rootCmd.Version = GetFullVersion()

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chrono/config.yaml)")
	rootCmd.PersistentFlags().String("api-url", "", "API server URL (overrides config file)")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug output")

	// Bind flags to viper
	viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory
		viper.AddConfigPath(home + "/.chrono")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// If a config file is found, read it in
	if err := viper.MergeInConfig(); err == nil {
		// Config file found and successfully parsed
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// Config file was found but there was an error
		cobra.CheckErr(err)
	}
}

// GetConfig loads and returns the configuration
func GetConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Apply API URL overrides from flags or config
	if viper.IsSet("api-url") {
		cfg.MCP.ServerURL = viper.GetString("api-url")
	}

	return cfg
}

// GetAPIClient returns an API client configured with the current settings
func GetAPIClient() *config.Config {
	cfg := GetConfig()
	return cfg
}
