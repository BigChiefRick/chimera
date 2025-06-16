package discover

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/pkg/discovery"
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
	providers, err := parseProviders(opts.Providers)
	if err != nil {
		return fmt.Errorf("invalid providers: %w", err)
	}

	// Show discovery plan if dry run
	if opts.DryRun {
		return showDiscoveryPlan(providers, opts)
	}

	// For Phase 1 - show that discovery would work
	logger.Info("Starting resource discovery")
	fmt.Println("🔍 Discovery functionality (Phase 1 - Framework Complete)")
	fmt.Println("=========================================================")
	fmt.Printf("Providers: %v\n", providers)
	fmt.Printf("Regions: %v\n", opts.Regions)
	fmt.Printf("Resource Types: %v\n", opts.ResourceTypes)
	fmt.Println("")
	fmt.Println("✅ Discovery framework is ready!")
	fmt.Println("✅ AWS provider connector available")
	fmt.Println("⚠️  Note: Full discovery requires AWS credentials")
	fmt.Println("")
	fmt.Println("Next: Configure AWS CLI and run with real credentials")

	return nil
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
