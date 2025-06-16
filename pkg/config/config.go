package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"github.com/sirupsen/logrus"
)

// Config represents the main configuration structure
type Config struct {
	// Global settings
	Debug        bool          `yaml:"debug" json:"debug"`
	Verbose      bool          `yaml:"verbose" json:"verbose"`
	OutputFormat string        `yaml:"output_format" json:"output_format"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout"`

	// Discovery settings
	Discovery DiscoveryConfig `yaml:"discovery" json:"discovery"`

	// Generation settings
	Generation GenerationConfig `yaml:"generation" json:"generation"`

	// Provider configurations
	Providers ProvidersConfig `yaml:"providers" json:"providers"`
}

// DiscoveryConfig contains discovery-specific configuration
type DiscoveryConfig struct {
	MaxConcurrency int             `yaml:"max_concurrency" json:"max_concurrency"`
	Steampipe      SteampipeConfig `yaml:"steampipe" json:"steampipe"`
}

// SteampipeConfig contains Steampipe-specific configuration
type SteampipeConfig struct {
	Host     string        `yaml:"host" json:"host"`
	Port     int           `yaml:"port" json:"port"`
	Database string        `yaml:"database" json:"database"`
	User     string        `yaml:"user" json:"user"`
	Password string        `yaml:"password" json:"password"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
}

// GenerationConfig contains generation-specific configuration
type GenerationConfig struct {
	OutputPath      string `yaml:"output_path" json:"output_path"`
	OrganizeByType  bool   `yaml:"organize_by_type" json:"organize_by_type"`
	IncludeState    bool   `yaml:"include_state" json:"include_state"`
	ValidateOutput  bool   `yaml:"validate_output" json:"validate_output"`
}

// ProvidersConfig contains provider-specific configurations
type ProvidersConfig struct {
	AWS    AWSConfig    `yaml:"aws" json:"aws"`
	Azure  AzureConfig  `yaml:"azure" json:"azure"`
	GCP    GCPConfig    `yaml:"gcp" json:"gcp"`
	VMware VMwareConfig `yaml:"vmware" json:"vmware"`
	KVM    KVMConfig    `yaml:"kvm" json:"kvm"`
}

// AWSConfig contains AWS-specific configuration
type AWSConfig struct {
	Regions []string `yaml:"regions" json:"regions"`
	Profile string   `yaml:"profile" json:"profile"`
}

// AzureConfig contains Azure-specific configuration
type AzureConfig struct {
	SubscriptionID string   `yaml:"subscription_id" json:"subscription_id"`
	TenantID       string   `yaml:"tenant_id" json:"tenant_id"`
	Locations      []string `yaml:"locations" json:"locations"`
}

// GCPConfig contains GCP-specific configuration
type GCPConfig struct {
	ProjectID string   `yaml:"project_id" json:"project_id"`
	Regions   []string `yaml:"regions" json:"regions"`
	Zones     []string `yaml:"zones" json:"zones"`
}

// VMwareConfig contains VMware vSphere configuration
type VMwareConfig struct {
	VCenterHost string `yaml:"vcenter_host" json:"vcenter_host"`
	Username    string `yaml:"username" json:"username"`
	Password    string `yaml:"password" json:"password"`
	Datacenter  string `yaml:"datacenter" json:"datacenter"`
}

// KVMConfig contains KVM/libvirt configuration
type KVMConfig struct {
	URI        string   `yaml:"uri" json:"uri"`
	Hosts      []string `yaml:"hosts" json:"hosts"`
	Username   string   `yaml:"username" json:"username"`
	KeyFile    string   `yaml:"key_file" json:"key_file"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Debug:        false,
		Verbose:      false,
		OutputFormat: "json",
		Timeout:      10 * time.Minute,
		Discovery: DiscoveryConfig{
			MaxConcurrency: 10,
			Steampipe: SteampipeConfig{
				Host:     "localhost",
				Port:     9193,
				Database: "steampipe",
				User:     "steampipe",
				Timeout:  30 * time.Second,
			},
		},
		Generation: GenerationConfig{
			OutputPath:     "./generated",
			OrganizeByType: true,
			IncludeState:   true,
			ValidateOutput: true,
		},
		Providers: ProvidersConfig{
			AWS: AWSConfig{
				Regions: []string{"us-east-1", "us-west-2"},
			},
			Azure: AzureConfig{
				Locations: []string{"East US", "West US 2"},
			},
			GCP: GCPConfig{
				Regions: []string{"us-central1", "us-east1"},
			},
		},
	}
}

// LoadConfig loads configuration from various sources
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Set up viper
	viper.SetConfigName(".chimera")
	viper.SetConfigType("yaml")
	
	// Add config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc/chimera")

	// Set environment variable prefix
	viper.SetEnvPrefix("CHIMERA")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is not an error - we'll use defaults
		logrus.Debug("No config file found, using defaults")
	} else {
		logrus.Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	// Unmarshal into our config struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// InitializeConfig creates a default configuration file
func InitializeConfig() error {
	config := DefaultConfig()
	
	// Determine config file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configPath := filepath.Join(homeDir, ".chimera.yaml")
	
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Save the default config
	if err := SaveConfig(config, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Configuration file created at: %s\n", configPath)
	fmt.Println("You can now edit this file to customize your settings.")
	
	return nil
}

// ValidateConfig validates the configuration
func ValidateConfig() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate global settings
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	validFormats := []string{"json", "yaml", "table"}
	validFormat := false
	for _, format := range validFormats {
		if config.OutputFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid output format: %s (valid: json, yaml, table)", config.OutputFormat)
	}

	// Validate discovery settings
	if config.Discovery.MaxConcurrency <= 0 {
		return fmt.Errorf("discovery max_concurrency must be greater than 0")
	}

	if config.Discovery.Steampipe.Port <= 0 || config.Discovery.Steampipe.Port > 65535 {
		return fmt.Errorf("steampipe port must be between 1 and 65535")
	}

	if config.Discovery.Steampipe.Timeout <= 0 {
		return fmt.Errorf("steampipe timeout must be greater than 0")
	}

	// Validate generation settings
	if config.Generation.OutputPath == "" {
		return fmt.Errorf("generation output_path cannot be empty")
	}

	fmt.Println("Configuration is valid!")
	return nil
}

// ShowConfig displays the current configuration
func ShowConfig() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Marshal to YAML for display
	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	fmt.Println("Current Configuration:")
	fmt.Println("=====================")
	fmt.Print(string(data))
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	if viper.ConfigFileUsed() != "" {
		return viper.ConfigFileUsed()
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".chimera.yaml"
	}

	return filepath.Join(homeDir, ".chimera.yaml")
}