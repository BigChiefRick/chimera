package discover

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/BigChiefRick/chimera/pkg/discovery"
	"github.com/BigChiefRick/chimera/pkg/discovery/providers"
	"github.com/BigChiefRick/chimera/pkg/discovery/steampipe"
)

// Options contains the discover command options
type Options struct {
	Providers      []string
	Regions        []string
	ResourceTypes  []string
	OutputPath     string
	OutputFormat   string
	MaxConcurrency int
	Timeout        time.Duration
	UseSteampipe   bool
	Filters        []string
	IncludeTags    []string
	ExcludeTags    []string
	Verbose        bool
	DryRun         bool
}

// NewDiscoverCommand creates the discover command
func NewDiscoverCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover infrastructure resources from cloud providers",
		Long: `Discover infrastructure resources from multiple cloud providers including
AWS, Azure, GCP, VMware vSphere, and KVM environments.

Examples:
  # Discover all resources from AWS in us-east-1
  chimera discover --provider aws --region us-east-1

  # Discover VPCs and subnets from multiple providers
  chimera discover --provider aws,azure --resource-type vpc,subnet

  # Use Steampipe for unified discovery
  chimera discover --steampipe --provider aws,azure,gcp

  # Apply filters to discovery
  chimera discover --provider aws --filter "name=prod-*" --filter "tag:Environment=production"

  # Output to file in YAML format
  chimera discover --provider aws --output resources.yaml --format yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscover(cmd.Context(), opts)
		},
	}

	// Provider flags
	cmd.Flags().StringSliceVar(&opts.Providers, "provider", []string{}, 
		"Cloud providers to discover from (aws,azure,gcp,vmware,kvm)")
	cmd.Flags().StringSliceVar(&opts.Regions, "region", []string{}, 
		"Regions to discover resources from")
	cmd.Flags().StringSliceVar(&opts.ResourceTypes, "resource-type", []string{}, 
		"Specific resource types to discover")

	// Output flags
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", 
		"Output file path (default: stdout)")
	cmd.Flags().StringVarP(&opts.OutputFormat, "format", "f", "json", 
		"Output format (json,yaml,table)")

	// Discovery flags
	cmd.Flags().IntVar(&opts.MaxConcurrency, "concurrency", 10, 
		"Maximum concurrent discovery operations")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 10*time.Minute, 
		"Discovery timeout")
	cmd.Flags().BoolVar(&opts.UseSteampipe, "steampipe", false, 
		"Use Steampipe for unified discovery")

	// Filtering flags
	cmd.Flags().StringSliceVar(&opts.Filters, "filter", []string{}, 
		"Resource filters (key=value or key:operator:value)")
	cmd.Flags().StringSliceVar(&opts.IncludeTags, "include-tag", []string{}, 
		"Include resources with these tags")
	cmd.Flags().StringSliceVar(&opts.ExcludeTags, "exclude-tag", []string{}, 
		"Exclude resources with these tags")

	// Behavior flags
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, 
		"Verbose output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, 
		"Show what would be discovered without actually discovering")

	// Required flags
	cmd.MarkFlagRequired("provider")

	return cmd
}

// runDiscover executes the discover command
func runDiscover(ctx context.Context, opts *Options) error {
	// Setup logging
	if opts.Verbose {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logger := logrus.WithField("command", "discover")
	logger.Info("Starting resource discovery")

	// Validate options
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Convert string providers to CloudProvider enum
	providers, err := parseProviders(opts.Providers)
	if err != nil {
		return fmt.Errorf("invalid providers: %w", err)
	}

	// Parse filters
	filters, err := parseFilters(opts.Filters)
	if err != nil {
		return fmt.Errorf("invalid filters: %w", err)
	}

	// Add tag filters
	filters = append(filters, parseTagFilters(opts.IncludeTags, opts.ExcludeTags)...)

	// Prepare discovery options
	discoveryOpts := discovery.DiscoveryOptions{
		Providers:      providers,
		Regions:        opts.Regions,
		ResourceTypes:  opts.ResourceTypes,
		MaxConcurrency: opts.MaxConcurrency,
		Timeout:        opts.Timeout,
		Filters:        filters,
	}

	// Show discovery plan if dry run
	if opts.DryRun {
		return showDiscoveryPlan(discoveryOpts)
	}

	// Create discovery engine
	engine, err := createDiscoveryEngine(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create discovery engine: %w", err)
	}

	// Validate credentials before discovery
	if err := engine.ValidateCredentials(ctx, providers); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	logger.Info("Credentials validated, starting discovery")

	// Execute discovery
	result, err := engine.Discover(ctx, discoveryOpts)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	// Log summary
	logger.WithFields(logrus.Fields{
		"resources_found": result.Metadata.ResourceCount,
		"providers":       len(providers),
		"duration":        result.Metadata.Duration,
		"errors":          result.Metadata.ErrorCount,
	}).Info("Discovery completed")

	// Output results
	return outputResults(result, opts)
}

// validateOptions validates the discover command options
func validateOptions(opts *Options) error {
	if len(opts.Providers) == 0 {
		return fmt.Errorf("at least one provider must be specified")
	}

	validFormats := []string{"json", "yaml", "table"}
	validFormat := false
	for _, format := range validFormats {
		if opts.OutputFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid output format: %s (valid: %s)", 
			opts.OutputFormat, strings.Join(validFormats, ","))
	}

	return nil
}

// parseProviders converts string slice to CloudProvider slice
func parseProviders(providerStrs []string) ([]discovery.CloudProvider, error) {
	var providers []discovery.CloudProvider
	
	for _, providerStr := range providerStrs {
		switch strings.ToLower(providerStr) {
		case "aws":
			providers = append(providers, discovery.AWS)
		case "azure":
			providers = append(providers, discovery.Azure)
		case "gcp":
			providers = append(providers, discovery.GCP)
		case "vmware":
			providers = append(providers, discovery.VMware)
		case "kvm":
			providers = append(providers, discovery.KVM)
		case "kubernetes", "k8s":
			providers = append(providers, discovery.Kubernetes)
		default:
			return nil, fmt.Errorf("unsupported provider: %s", providerStr)
		}
	}
	
	return providers, nil
}

// parseFilters parses filter strings into Filter objects
func parseFilters(filterStrs []string) ([]discovery.Filter, error) {
	var filters []discovery.Filter
	
	for _, filterStr := range filterStrs {
		filter, err := parseFilter(filterStr)
		if err != nil {
			return nil, fmt.Errorf("invalid filter '%s': %w", filterStr, err)
		}
		filters = append(filters, filter)
	}
	
	return filters, nil
}

// parseFilter parses a single filter string
func parseFilter(filterStr string) (discovery.Filter, error) {
	// Support formats:
	// - key=value
	// - key:operator:value
	// - tag:key=value
	
	if strings.Contains(filterStr, ":") {
		parts := strings.SplitN(filterStr, ":", 3)
		if len(parts) == 3 {
			return discovery.Filter{
				Field:    parts[0],
				Operator: parts[1],
				Value:    parts[2],
			}, nil
		} else if len(parts) == 2 && strings.Contains(parts[1], "=") {
			// Handle tag:key=value
			tagParts := strings.SplitN(parts[1], "=", 2)
			return discovery.Filter{
				Field:    parts[0] + "." + tagParts[0],
				Operator: "eq",
				Value:    tagParts[1],
			}, nil
		}
	}
	
	if strings.Contains(filterStr, "=") {
		parts := strings.SplitN(filterStr, "=", 2)
		return discovery.Filter{
			Field:    parts[0],
			Operator: "eq",
			Value:    parts[1],
		}, nil
	}
	
	return discovery.Filter{}, fmt.Errorf("invalid filter format")
}

// parseTagFilters creates filters for include/exclude tags
func parseTagFilters(includeTags, excludeTags []string) []discovery.Filter {
	var filters []discovery.Filter
	
	for _, tag := range includeTags {
		if strings.Contains(tag, "=") {
			parts := strings.SplitN(tag, "=", 2)
			filters = append(filters, discovery.Filter{
				Field:    "tags." + parts[0],
				Operator: "eq",
				Value:    parts[1],
			})
		} else {
			// Tag key exists
			filters = append(filters, discovery.Filter{
				Field:    "tags." + tag,
				Operator: "exists",
				Value:    true,
			})
		}
	}
	
	for _, tag := range excludeTags {
		if strings.Contains(tag, "=") {
			parts := strings.SplitN(tag, "=", 2)
			filters = append(filters, discovery.Filter{
				Field:    "tags." + parts[0],
				Operator: "ne",
				Value:    parts[1],
			})
		} else {
			// Tag key does not exist
			filters = append(filters, discovery.Filter{
				Field:    "tags." + tag,
				Operator: "exists",
				Value:    false,
			})
		}
	}
	
	return filters
}

// createDiscoveryEngine creates and configures the discovery engine
func createDiscoveryEngine(ctx context.Context, opts *Options) (discovery.DiscoveryEngine, error) {
	engineConfig := discovery.EngineConfig{
		MaxConcurrency: opts.MaxConcurrency,
		Timeout:        opts.Timeout,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
	}

	var steampipeConnector discovery.SteampipeConnector
	if opts.UseSteampipe {
		// Get Steampipe configuration from viper
		steampipeConfig := steampipe.Config{
			Host:     viper.GetString("discovery.steampipe.host"),
			Port:     viper.GetInt("discovery.steampipe.port"),
			Database: viper.GetString("discovery.steampipe.database"),
			User:     viper.GetString("discovery.steampipe.user"),
			Timeout:  viper.GetDuration("discovery.steampipe.timeout"),
		}
		
		// Use defaults if not configured
		if steampipeConfig.Host == "" {
			steampipeConfig.Host = "localhost"
		}
		if steampipeConfig.Port == 0 {
			steampipeConfig.Port = 9193
		}
		if steampipeConfig.Database == "" {
			steampipeConfig.Database = "steampipe"
		}
		if steampipeConfig.User == "" {
			steampipeConfig.User = "steampipe"
		}
		if steampipeConfig.Timeout == 0 {
			steampipeConfig.Timeout = 30 * time.Second
		}
		
		steampipeConnector = steampipe.NewConnector(steampipeConfig)
	}

	engine := discovery.NewEngine(engineConfig, steampipeConnector)

	// Register provider connectors
	for _, providerStr := range opts.Providers {
		switch strings.ToLower(providerStr) {
		case "aws":
			// Register AWS connector for each region
			regions := opts.Regions
			if len(regions) == 0 {
				regions = []string{"us-east-1"} // Default region
			}
			
			// Create one connector that handles all regions
			awsConnector, err := providers.NewAWSConnector(ctx, regions[0])
			if err != nil {
				return nil, fmt.Errorf("failed to create AWS connector: %w", err)
			}
			engine.RegisterConnector(awsConnector)
			
		case "azure":
			// TODO: Implement Azure connector
			logrus.Warn("Azure connector not yet implemented")
		case "gcp":
			// TODO: Implement GCP connector
			logrus.Warn("GCP connector not yet implemented")
		case "vmware":
			// TODO: Implement VMware connector
			logrus.Warn("VMware connector not yet implemented")
		case "kvm":
			// TODO: Implement KVM connector
			logrus.Warn("KVM connector not yet implemented")
		}
	}

	return engine, nil
}

// showDiscoveryPlan shows what would be discovered in a dry run
func showDiscoveryPlan(opts discovery.DiscoveryOptions) error {
	fmt.Println("🔍 Discovery Plan:")
	fmt.Println("=================")
	fmt.Printf("Providers: %v\n", opts.Providers)
	
	if len(opts.Regions) > 0 {
		fmt.Printf("Regions: %v\n", opts.Regions)
	} else {
		fmt.Println("Regions: all available")
	}
	
	if len(opts.ResourceTypes) > 0 {
		fmt.Printf("Resource Types: %v\n", opts.ResourceTypes)
	} else {
		fmt.Println("Resource Types: all supported")
	}
	
	if len(opts.Filters) > 0 {
		fmt.Println("Filters:")
		for _, filter := range opts.Filters {
			fmt.Printf("  %s %s %v\n", filter.Field, filter.Operator, filter.Value)
		}
	}
	
	fmt.Printf("Max Concurrency: %d\n", opts.MaxConcurrency)
	fmt.Printf("Timeout: %v\n", opts.Timeout)
	
	fmt.Println("\n✅ This is what would be discovered.")
	fmt.Println("Remove --dry-run to execute actual discovery.")
	
	return nil
}

// outputResults outputs the discovery results in the specified format
func outputResults(result *discovery.DiscoveryResult, opts *Options) error {
	var output []byte
	var err error

	switch opts.OutputFormat {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(result)
	case "table":
		return outputTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", opts.OutputFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	// Write to file or stdout
	if opts.OutputPath != "" {
		if err := os.WriteFile(opts.OutputPath, output, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Results written to %s\n", opts.OutputPath)
	} else {
		fmt.Print(string(output))
	}

	return nil
}

// outputTable outputs results in a table format
func outputTable(result *discovery.DiscoveryResult) error {
	// Print summary
	fmt.Printf("🔍 Discovery Summary:\n")
	fmt.Printf("====================\n")
	fmt.Printf("  Resources Found: %d\n", result.Metadata.ResourceCount)
	fmt.Printf("  Duration: %v\n", result.Metadata.Duration)
	fmt.Printf("  Errors: %d\n", result.Metadata.ErrorCount)
	fmt.Printf("\n")

	// Print provider statistics
	if len(result.Metadata.ProviderStats) > 0 {
		fmt.Printf("Provider Statistics:\n")
		for provider, count := range result.Metadata.ProviderStats {
			fmt.Printf("  %s: %d resources\n", provider, count)
		}
		fmt.Printf("\n")
	}

	// Print resources table
	if len(result.Resources) > 0 {
		fmt.Printf("%-50s %-20s %-15s %-15s %-20s\n", 
			"ID", "Name", "Type", "Provider", "Region")
		fmt.Printf("%s\n", strings.Repeat("-", 120))
		
		for _, resource := range result.Resources {
			name := resource.Name
			if name == "" {
				name = "-"
			}
			region := resource.Region
			if region == "" {
				region = "-"
			}
			
			fmt.Printf("%-50s %-20s %-15s %-15s %-20s\n",
				truncateString(resource.ID, 50),
				truncateString(name, 20),
				truncateString(resource.Type, 15),
				string(resource.Provider),
				region)
		}
	}

	// Print errors if any
	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  [%s] %s\n", err.Provider, err.Message)
		}
	}

	return nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
