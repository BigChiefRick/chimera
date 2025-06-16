package discover

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/pkg/discovery"
	"github.com/BigChiefRick/chimera/pkg/discovery/providers"
)

// Options contains the discover command options
type Options struct {
	Providers        []string
	Regions          []string
	ResourceTypes    []string
	OutputPath       string
	OutputFormat     string
	MaxConcurrency   int
	Timeout          time.Duration
	Verbose          bool
	DryRun           bool
	ForceReal        bool
	// Cloud-specific options
	AWSProfile       string
	AzureSubscription string
	GCPProject       string
}

// NewDiscoverCommand creates the discover command
func NewDiscoverCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover infrastructure resources from cloud providers",
		Long: `Discover infrastructure resources from multiple cloud providers.
		
Phase 2: Multi-cloud support with AWS, Azure, and GCP
Supports real discovery across all major cloud platforms.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscover(cmd.Context(), opts)
		},
	}

	// Provider flags
	cmd.Flags().StringSliceVar(&opts.Providers, "provider", []string{}, 
		"Cloud providers to discover from (aws,azure,gcp)")
	cmd.Flags().StringSliceVar(&opts.Regions, "region", []string{}, 
		"Regions to discover resources from")
	cmd.Flags().StringSliceVar(&opts.ResourceTypes, "resource-type", []string{}, 
		"Specific resource types to discover")

	// Cloud-specific flags
	cmd.Flags().StringVar(&opts.AWSProfile, "aws-profile", "", 
		"AWS profile to use (overrides default)")
	cmd.Flags().StringVar(&opts.AzureSubscription, "azure-subscription", "", 
		"Azure subscription ID (required for Azure)")
	cmd.Flags().StringVar(&opts.GCPProject, "gcp-project", "", 
		"GCP project ID (required for GCP)")

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

	// Behavior flags
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, 
		"Verbose output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, 
		"Show what would be discovered without actually discovering")
	cmd.Flags().BoolVar(&opts.ForceReal, "real", false, 
		"Force real discovery (bypass credential check)")

	// Required flags
	cmd.MarkFlagRequired("provider")

	return cmd
}

// runDiscover executes the discover command
func runDiscover(ctx context.Context, opts *Options) error {
	if opts.Verbose {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logger := logrus.WithField("command", "discover")

	// Validate options
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Parse providers
	providerTypes, err := parseProviders(opts.Providers)
	if err != nil {
		return fmt.Errorf("invalid providers: %w", err)
	}

	// Show discovery plan if dry run
	if opts.DryRun {
		return showDiscoveryPlan(providerTypes, opts)
	}

	// Always attempt real discovery in Phase 2
	return performMultiCloudDiscovery(ctx, opts, providerTypes, logger)
}

// performMultiCloudDiscovery performs discovery across multiple cloud providers
func performMultiCloudDiscovery(ctx context.Context, opts *Options, providerTypes []discovery.CloudProvider, logger *logrus.Entry) error {
	fmt.Println("🔍 Multi-Cloud Infrastructure Discovery (Phase 2)")
	fmt.Println("================================================")

	var allResources []discovery.Resource
	var allErrors []discovery.DiscoveryError
	providerStats := make(map[string]int)
	startTime := time.Now()

	// Discover resources from each provider
	for _, provider := range providerTypes {
		fmt.Printf("\n🔍 Discovering %s resources...\n", strings.ToUpper(string(provider)))
		
		resources, err := discoverProviderResources(ctx, provider, opts)
		if err != nil {
			logger.Errorf("Failed to discover %s resources: %v", provider, err)
			allErrors = append(allErrors, discovery.DiscoveryError{
				Provider: provider,
				Message:  err.Error(),
			})
			continue
		}

		providerStats[string(provider)] = len(resources)
		allResources = append(allResources, resources...)
		
		fmt.Printf("✅ Found %d %s resources\n", len(resources), provider)
	}

	duration := time.Since(startTime)

	// Create combined result
	result := &discovery.DiscoveryResult{
		Resources: allResources,
		Errors:    allErrors,
		Metadata: discovery.DiscoveryMetadata{
			StartTime:     startTime,
			EndTime:       time.Now(),
			Duration:      duration,
			ResourceCount: len(allResources),
			ProviderStats: providerStats,
			ErrorCount:    len(allErrors),
		},
	}

	// Output results
	fmt.Printf("\n🎉 Multi-Cloud Discovery Complete!\n")
	fmt.Printf("Total resources found: %d\n", len(allResources))
	fmt.Printf("Discovery duration: %v\n", duration)
	
	if len(allErrors) > 0 {
		fmt.Printf("⚠️  Encountered %d errors during discovery\n", len(allErrors))
	}

	return outputResults(result, opts)
}

// discoverProviderResources discovers resources from a specific cloud provider
func discoverProviderResources(ctx context.Context, provider discovery.CloudProvider, opts *Options) ([]discovery.Resource, error) {
	switch provider {
	case discovery.AWS:
		return discoverAWSResources(ctx, opts)
	case discovery.Azure:
		return discoverAzureResources(ctx, opts)
	case discovery.GCP:
		return discoverGCPResources(ctx, opts)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// discoverAWSResources discovers AWS resources
func discoverAWSResources(ctx context.Context, opts *Options) ([]discovery.Resource, error) {
	// Use first region or default
	region := "us-east-1"
	if len(opts.Regions) > 0 {
		region = opts.Regions[0]
	}

	// Set AWS profile if specified
	if opts.AWSProfile != "" {
		os.Setenv("AWS_PROFILE", opts.AWSProfile)
	}

	// Create AWS connector
	awsConnector, err := providers.NewAWSConnector(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS connector: %w", err)
	}

	// Validate credentials
	if err := awsConnector.ValidateCredentials(ctx); err != nil {
		return nil, fmt.Errorf("AWS credential validation failed: %w", err)
	}

	// Prepare discovery options
	providerOpts := discovery.ProviderDiscoveryOptions{
		Regions:       opts.Regions,
		ResourceTypes: opts.ResourceTypes,
	}

	return awsConnector.Discover(ctx, providerOpts)
}

// discoverAzureResources discovers Azure resources
func discoverAzureResources(ctx context.Context, opts *Options) ([]discovery.Resource, error) {
	if opts.AzureSubscription == "" {
		return nil, fmt.Errorf("Azure subscription ID is required (use --azure-subscription)")
	}

	// Create Azure connector
	azureConnector, err := providers.NewAzureConnector(ctx, opts.AzureSubscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure connector: %w", err)
	}

	// Validate credentials
	if err := azureConnector.ValidateCredentials(ctx); err != nil {
		return nil, fmt.Errorf("Azure credential validation failed: %w", err)
	}

	// Prepare discovery options
	providerOpts := discovery.ProviderDiscoveryOptions{
		Regions:       opts.Regions,
		ResourceTypes: opts.ResourceTypes,
	}

	return azureConnector.Discover(ctx, providerOpts)
}

// discoverGCPResources discovers GCP resources
func discoverGCPResources(ctx context.Context, opts *Options) ([]discovery.Resource, error) {
	if opts.GCPProject == "" {
		return nil, fmt.Errorf("GCP project ID is required (use --gcp-project)")
	}

	// Create GCP connector
	gcpConnector, err := providers.NewGCPConnector(ctx, opts.GCPProject)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP connector: %w", err)
	}

	// Validate credentials
	if err := gcpConnector.ValidateCredentials(ctx); err != nil {
		return nil, fmt.Errorf("GCP credential validation failed: %w", err)
	}

	// Prepare discovery options
	providerOpts := discovery.ProviderDiscoveryOptions{
		Regions:       opts.Regions,
		ResourceTypes: opts.ResourceTypes,
	}

	return gcpConnector.Discover(ctx, providerOpts)
}

// outputResults outputs the discovery results
func outputResults(result *discovery.DiscoveryResult, opts *Options) error {
	if len(result.Resources) == 0 {
		fmt.Println("\n🔍 Discovery Complete - No resources found")
		fmt.Printf("   Providers: %v\n", opts.Providers)
		fmt.Printf("   Regions: %v\n", opts.Regions)
		return nil
	}

	fmt.Printf("\n📊 Resource Summary by Provider:\n")
	for provider, count := range result.Metadata.ProviderStats {
		fmt.Printf("   %s: %d resources\n", strings.ToUpper(provider), count)
	}
	fmt.Println()

	switch opts.OutputFormat {
	case "json":
		return outputJSON(result, opts.OutputPath)
	case "yaml":
		return outputYAML(result, opts.OutputPath)
	case "table":
		return outputTable(result.Resources)
	default:
		return outputJSON(result, opts.OutputPath)
	}
}

// outputJSON outputs results as JSON
func outputJSON(result *discovery.DiscoveryResult, outputPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if outputPath != "" {
		err = os.WriteFile(outputPath, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("✅ Results written to: %s\n", outputPath)
	} else {
		fmt.Println(string(data))
	}

	return nil
}

// outputYAML outputs results as YAML (placeholder)
func outputYAML(result *discovery.DiscoveryResult, outputPath string) error {
	// For now, fall back to JSON
	fmt.Println("YAML output not yet implemented, using JSON:")
	return outputJSON(result, outputPath)
}

// outputTable outputs results as a table
func outputTable(resources []discovery.Resource) error {
	if len(resources) == 0 {
		fmt.Println("No resources to display")
		return nil
	}

	fmt.Printf("%-10s %-25s %-30s %-25s %-15s %-15s\n", "PROVIDER", "NAME", "TYPE", "ID", "REGION", "ZONE")
	fmt.Println(strings.Repeat("-", 130))

	for _, resource := range resources {
		name := resource.Name
		if name == "" {
			name = "<unnamed>"
		}
		
		provider := strings.ToUpper(string(resource.Provider))
		
		fmt.Printf("%-10s %-25s %-30s %-25s %-15s %-15s\n",
			truncate(provider, 10),
			truncate(name, 25),
			truncate(resource.Type, 30),
			truncate(resource.ID, 25),
			truncate(resource.Region, 15),
			truncate(resource.Zone, 15),
		)
	}

	fmt.Printf("\nTotal: %d resources\n", len(resources))
	return nil
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// validateOptions validates the discover command options
func validateOptions(opts *Options) error {
	if len(opts.Providers) == 0 {
		return fmt.Errorf("at least one provider must be specified")
	}

	// Validate cloud-specific requirements
	for _, provider := range opts.Providers {
		switch strings.ToLower(provider) {
		case "azure":
			if opts.AzureSubscription == "" {
				return fmt.Errorf("Azure subscription ID is required (use --azure-subscription)")
			}
		case "gcp":
			if opts.GCPProject == "" {
				return fmt.Errorf("GCP project ID is required (use --gcp-project)")
			}
		}
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
			return nil, fmt.Errorf("vmware provider not yet implemented (Phase 3)")
		case "kvm":
			return nil, fmt.Errorf("kvm provider not yet implemented (Phase 3)")
		default:
			return nil, fmt.Errorf("unsupported provider: %s", providerStr)
		}
	}
	
	return providers, nil
}

// showDiscoveryPlan shows what would be discovered in a dry run
func showDiscoveryPlan(providers []discovery.CloudProvider, opts *Options) error {
	fmt.Println("🔍 Multi-Cloud Discovery Plan (Phase 2):")
	fmt.Println("========================================")
	fmt.Printf("Providers: %v\n", providers)
	
	if len(opts.Regions) > 0 {
		fmt.Printf("Regions: %v\n", opts.Regions)
	} else {
		fmt.Println("Regions: all available per provider")
	}
	
	if len(opts.ResourceTypes) > 0 {
		fmt.Printf("Resource Types: %v\n", opts.ResourceTypes)
	} else {
		fmt.Println("Resource Types: all supported per provider")
	}
	
	fmt.Printf("Max Concurrency: %d\n", opts.MaxConcurrency)
	fmt.Printf("Timeout: %v\n", opts.Timeout)
	fmt.Printf("Output Format: %s\n", opts.OutputFormat)
	
	if opts.OutputPath != "" {
		fmt.Printf("Output File: %s\n", opts.OutputPath)
	}

	// Show provider-specific information
	fmt.Println("\nProvider-Specific Details:")
	for _, provider := range providers {
		switch provider {
		case discovery.AWS:
			fmt.Printf("  AWS: Profile=%s, Regions=%v\n", 
				opts.AWSProfile, opts.Regions)
		case discovery.Azure:
			fmt.Printf("  Azure: Subscription=%s, Regions=%v\n", 
				opts.AzureSubscription, opts.Regions)
		case discovery.GCP:
			fmt.Printf("  GCP: Project=%s, Regions=%v\n", 
				opts.GCPProject, opts.Regions)
		}
	}
	
	fmt.Println("\n✅ This is what would be discovered across all providers.")
	fmt.Println("Remove --dry-run to execute actual multi-cloud discovery.")
	
	return nil
}
