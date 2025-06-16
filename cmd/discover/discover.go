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
	Providers      []string
	Regions        []string
	ResourceTypes  []string
	OutputPath     string
	OutputFormat   string
	MaxConcurrency int
	Timeout        time.Duration
	Verbose        bool
	DryRun         bool
	ForceReal      bool // Add this to force real discovery
}

// NewDiscoverCommand creates the discover command
func NewDiscoverCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover infrastructure resources from cloud providers",
		Long: `Discover infrastructure resources from multiple cloud providers.
		
Phase 1: AWS support with basic EC2 discovery
Phase 2: Full multi-cloud support with Azure, GCP, VMware, KVM`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscover(cmd.Context(), opts)
		},
	}

	// Provider flags
	cmd.Flags().StringSliceVar(&opts.Providers, "provider", []string{}, 
		"Cloud providers to discover from (aws)")
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

	// FIXED: Always try real discovery first, fall back to demo only if it fails
	if opts.ForceReal || shouldAttemptRealDiscovery() {
		// Try real discovery
		err := performRealDiscovery(ctx, opts, providerTypes, logger)
		if err == nil {
			return nil // Success!
		}
		
		// If real discovery failed, show the error and continue to demo
		fmt.Printf("⚠️  Real discovery failed: %v\n", err)
		fmt.Println("Falling back to Phase 1 demo...\n")
	}

	// Show Phase 1 framework demo
	return showFrameworkDemo(providerTypes, opts)
}

// shouldAttemptRealDiscovery determines if we should attempt real discovery
func shouldAttemptRealDiscovery() bool {
	// Always attempt real discovery - let the AWS SDK handle credential detection
	return true
}

// performRealDiscovery performs the actual AWS discovery
func performRealDiscovery(ctx context.Context, opts *Options, providerTypes []discovery.CloudProvider, logger *logrus.Entry) error {
	fmt.Println("🔍 Attempting Real AWS Discovery")
	fmt.Println("================================")

	// Handle only AWS for now
	if len(providerTypes) != 1 || providerTypes[0] != discovery.AWS {
		return fmt.Errorf("only AWS provider supported for real discovery in Phase 1")
	}

	// Use first region or default
	region := "us-east-1"
	if len(opts.Regions) > 0 {
		region = opts.Regions[0]
	}

	fmt.Printf("🔍 Target region: %s\n", region)

	// Create AWS connector
	awsConnector, err := providers.NewAWSConnector(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to create AWS connector: %w", err)
	}

	// Validate credentials
	fmt.Println("🔑 Validating AWS credentials...")
	if err := awsConnector.ValidateCredentials(ctx); err != nil {
		return fmt.Errorf("AWS credential validation failed: %w", err)
	}

	fmt.Println("✅ AWS credentials validated successfully!")

	// Prepare provider discovery options
	providerOpts := discovery.ProviderDiscoveryOptions{
		Regions:       opts.Regions,
		ResourceTypes: opts.ResourceTypes,
	}

	// Perform discovery
	fmt.Println("🔍 Scanning for AWS resources...")
	startTime := time.Now()
	
	resources, err := awsConnector.Discover(ctx, providerOpts)
	if err != nil {
		return fmt.Errorf("AWS discovery failed: %w", err)
	}

	duration := time.Since(startTime)

	// Create result structure
	result := &discovery.DiscoveryResult{
		Resources: resources,
		Metadata: discovery.DiscoveryMetadata{
			StartTime:     startTime,
			EndTime:       time.Now(),
			Duration:      duration,
			ResourceCount: len(resources),
			ProviderStats: map[string]int{
				"aws": len(resources),
			},
		},
	}

	// Output results
	return outputResults(result, opts)
}

// showFrameworkDemo shows the Phase 1 framework demo
func showFrameworkDemo(providers []discovery.CloudProvider, opts *Options) error {
	fmt.Println("🔍 Discovery functionality (Phase 1 - Framework Complete)")
	fmt.Println("=========================================================")
	fmt.Printf("Providers: %v\n", providers)
	fmt.Printf("Regions: %v\n", opts.Regions)
	fmt.Printf("Resource Types: %v\n", opts.ResourceTypes)
	fmt.Println("")
	fmt.Println("✅ Discovery framework is ready!")
	fmt.Println("✅ AWS provider connector available")
	fmt.Println("⚠️  Note: Real discovery requires AWS credentials")
	fmt.Println("")
	fmt.Println("To enable real discovery:")
	fmt.Println("1. Configure AWS CLI: aws configure")
	fmt.Println("2. Or set environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
	fmt.Println("3. Add --real flag to force real discovery attempt")

	return nil
}

// outputResults outputs the discovery results
func outputResults(result *discovery.DiscoveryResult, opts *Options) error {
	if len(result.Resources) == 0 {
		fmt.Println("🔍 Discovery Complete - No resources found")
		fmt.Printf("   Scanned regions: %v\n", opts.Regions)
		fmt.Printf("   Resource types: %v\n", opts.ResourceTypes)
		fmt.Println("   This could mean:")
		fmt.Println("   • No resources exist in specified regions")
		fmt.Println("   • Insufficient permissions")
		fmt.Println("   • Region has no resources of specified types")
		return nil
	}

	fmt.Printf("🎉 Discovery Complete! Found %d resources\n", len(result.Resources))
	fmt.Printf("   Duration: %v\n", result.Metadata.Duration)
	fmt.Println("")

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

	fmt.Printf("%-25s %-20s %-25s %-15s %-15s\n", "NAME", "TYPE", "ID", "REGION", "ZONE")
	fmt.Println(strings.Repeat("-", 100))

	for _, resource := range resources {
		name := resource.Name
		if name == "" {
			name = "<unnamed>"
		}
		
		fmt.Printf("%-25s %-20s %-25s %-15s %-15s\n",
			truncate(name, 25),
			truncate(resource.Type, 20),
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
			return nil, fmt.Errorf("azure provider not yet implemented (Phase 2)")
		case "gcp":
			return nil, fmt.Errorf("gcp provider not yet implemented (Phase 2)")
		case "vmware":
			return nil, fmt.Errorf("vmware provider not yet implemented (Phase 2)")
		case "kvm":
			return nil, fmt.Errorf("kvm provider not yet implemented (Phase 2)")
		default:
			return nil, fmt.Errorf("unsupported provider: %s", providerStr)
		}
	}
	
	return providers, nil
}

// showDiscoveryPlan shows what would be discovered in a dry run
func showDiscoveryPlan(providers []discovery.CloudProvider, opts *Options) error {
	fmt.Println("🔍 Discovery Plan:")
	fmt.Println("=================")
	fmt.Printf("Providers: %v\n", providers)
	
	if len(opts.Regions) > 0 {
		fmt.Printf("Regions: %v\n", opts.Regions)
	} else {
		fmt.Println("Regions: all available")
	}
	
	if len(opts.ResourceTypes) > 0 {
		fmt.Printf("Resource Types: %v\n", opts.ResourceTypes)
	} else {
		fmt.Println("Resource Types: vpc, subnet, security_group, instance")
	}
	
	fmt.Printf("Max Concurrency: %d\n", opts.MaxConcurrency)
	fmt.Printf("Timeout: %v\n", opts.Timeout)
	fmt.Printf("Output Format: %s\n", opts.OutputFormat)
	
	if opts.OutputPath != "" {
		fmt.Printf("Output File: %s\n", opts.OutputPath)
	}
	
	fmt.Println("\n✅ This is what would be discovered.")
	fmt.Println("Remove --dry-run to execute actual discovery.")
	
	return nil
}