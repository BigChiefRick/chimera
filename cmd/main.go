package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/cmd/discover"
	"github.com/BigChiefRick/chimera/cmd/generate"
	"github.com/BigChiefRick/chimera/pkg/config"
)

var (
	cfgFile    string
	verbose    bool
	debug      bool
	outputJSON bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chimera",
	Short: "Multi-cloud infrastructure discovery and IaC generation tool",
	Long: `Chimera connects to multiple cloud and virtualization environments,
discovers infrastructure resources, and generates Infrastructure as Code templates.

Supported Platforms:
  • Amazon Web Services (AWS)
  • Microsoft Azure  
  • Google Cloud Platform (GCP)
  • VMware vSphere
  • KVM/libvirt

Supported IaC Outputs:
  • Terraform (.tf)
  • Pulumi
  • AWS CloudFormation
  • Azure ARM Templates`,
	Version: "0.1.0-alpha",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chimera.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug output")
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "output in JSON format")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))

	// Add subcommands
	rootCmd.AddCommand(discover.NewDiscoverCommand())
	rootCmd.AddCommand(generate.NewGenerateCommand())
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newConfigCommand())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".chimera" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".chimera")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

func setupLogging() {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if verbose {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	if outputJSON {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Chimera",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Chimera v%s\n", rootCmd.Version)
		},
	}
}

func newConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Chimera configuration",
		Long:  "Configure Chimera with cloud provider credentials and settings",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize Chimera configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.InitializeConfig(); err != nil {
				logrus.Fatalf("Failed to initialize config: %v", err)
			}
			fmt.Println("Configuration initialized successfully!")
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate Chimera configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.ValidateConfig(); err != nil {
				logrus.Fatalf("Configuration validation failed: %v", err)
			}
			fmt.Println("Configuration is valid!")
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			config.ShowConfig()
		},
	})

	return configCmd
}