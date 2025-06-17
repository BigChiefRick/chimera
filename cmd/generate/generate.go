package generate

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
	"github.com/BigChiefRick/chimera/pkg/generation"
	"github.com/BigChiefRick/chimera/pkg/generation/mappers"
	"github.com/BigChiefRick/chimera/pkg/generation/terraform"
)

// Options contains the generate command options
type Options struct {
	InputPath        string
	OutputPath       string
	Format           string
	Organization     string
	Providers        []string
	Verbose          bool
	DryRun           bool
	Timeout          time.Duration
	
	// Generation-specific options
	IncludeProvider  bool
	IncludeState     bool
	GenerateModules  bool
	ValidateOutput   bool
	SingleFile       bool
	CompactOutput    bool
	
	// Filtering options
	ExcludeResources []string
	IncludeResources []string
}

// NewGenerateCommand creates the generate command
func NewGenerateCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Infrastructure as Code from discovered resources",
		Long: `Generate Infrastructure as Code templates from previously discovered
infrastructure resources.

Phase 3: Full IaC generation with Terraform, multi-cloud support, and advanced features.

Examples:
  # Generate Terraform from discovery output
  chimera generate --input resources.json --format terraform --output ./infrastructure
  
  # Generate with specific organization
  chimera generate --input resources.json --organize-by provider --output ./infra
  
  # Generate for specific providers only
  chimera generate --input resources.json --provider aws --provider azure
  
  # Preview generation (dry-run)
  chimera generate --input resources.json --format terraform --dry-run
  
  # Generate with modules
  chimera generate --input resources.json --generate-modules --output ./modules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context(), opts)
		},
	}

	// Input/Output flags
	cmd.Flags().StringVarP(&opts.InputPath, "input", "i", "", 
		"Input file with discovered resources (required)")
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "./generated", 
		"Output directory for generated IaC files")

	// Format and organization flags
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "terraform", 
		"IaC format (terraform,terraform-json,pulumi)")
	cmd.Flags().StringVar(&opts.Organization, "organize-by", "provider", 
		"Organization method (provider,service,region,resource_type,flat)")

	// Provider filtering
	cmd.Flags().StringSliceVar(&opts.Providers, "provider", []string{}, 
		"Generate for specific providers only (aws,azure,gcp)")

	// Generation flags
	cmd.Flags().BoolVar(&opts.IncludeProvider, "include-provider", true, 
		"Include provider configuration files")
	cmd.Flags().BoolVar(&opts.IncludeState, "include-state", false, 
		"Include state configuration")
	cmd.Flags().BoolVar(&opts.GenerateModules, "generate-modules", false, 
		"Generate as Terraform modules")
	cmd.Flags().BoolVar(&opts.ValidateOutput, "validate", true, 
		"Validate generated output")
	cmd.Flags().BoolVar(&opts.SingleFile, "single-file", false, 
		"Generate everything in a single file")
	cmd.Flags().BoolVar(&opts.CompactOutput, "compact", false, 
		"Generate compact output")

	// Filtering flags
	cmd.Flags().StringSliceVar(&opts.ExcludeResources, "exclude", []string{}, 
		"Exclude specific resources (by type or ID)")
	cmd.Flags().StringSliceVar(&opts.IncludeResources, "include", []string{}, 
		"Include only specific resources (by type or ID)")

	// Behavior flags
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, 
		"Verbose output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, 
		"Show what would be generated without creating files")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 5*time.Minute, 
		"Generation timeout")

	// Mark required flags
	cmd.MarkFlagRequired("input")

	return cmd
}

// runGenerate executes the generate command
func runGenerate(ctx context.Context, opts *Options) error {
	if opts.Verbose {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logger := logrus.WithField("command", "generate")

	// Validate options
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	logger.Info("Starting IaC generation")

	// Load discovered resources
	resources, err := loadDiscoveredResources(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to load discovered resources: %w", err)
	}

	logger.Infof("Loaded %d discovered resources", len(resources))

	// Filter by provider if specified
	if len(opts.Providers) > 0 {
		resources = filterByProviders(resources, opts.Providers)
		logger.Infof("Filtered to %d resources for specified providers", len(resources))
	}

	if len(resources) == 0 {
		return fmt.Errorf("no resources available for generation")
	}

	// Create generation engine
	engine, err := createGenerationEngine(opts)
	if err != nil {
		return fmt.Errorf("failed to create generation engine: %w", err)
	}

	// Prepare generation options
	genOpts := generation.GenerationOptions{
		Resources:        resources,
		Format:           generation.IaCFormat(opts.Format),
		OutputPath:       opts.OutputPath,
		Organization:     generation.OrganizationPattern(opts.Organization),
		IncludeProvider:  opts.IncludeProvider,
		IncludeState:     opts.IncludeState,
		GenerateModules:  opts.GenerateModules,
		ValidateOutput:   opts.ValidateOutput,
		SingleFile:       opts.SingleFile,
		CompactOutput:    opts.CompactOutput,
		ExcludeResources: opts.ExcludeResources,
		IncludeResources: opts.IncludeResources,
		Timeout:          opts.Timeout,
	}

	// Show generation preview if dry run
	if opts.DryRun {
		return showGenerationPreview(ctx, engine, genOpts, logger)
	}

	// Perform actual generation
	return performGeneration(ctx, engine, genOpts, opts, logger)
}

// loadDiscoveredResources loads resources from discovery output file
func loadDiscoveredResources(inputPath string) ([]discovery.Resource, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	// Try to parse as DiscoveryResult first
	var discoveryResult discovery.DiscoveryResult
	if err := json.Unmarshal(data, &discoveryResult); err == nil {
		return discoveryResult.Resources, nil
	}

	// Try to parse as raw resource array
	var resources []discovery.Resource
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse input file as discovery result or resource array: %w", err)
	}

	return resources, nil
}

// filterByProviders filters resources by specified providers
func filterByProviders(resources []discovery.Resource, providers []string) []discovery.Resource {
	providerMap := make(map[string]bool)
	for _, provider := range providers {
		providerMap[strings.ToLower(provider)] = true
	}

	var filtered []discovery.Resource
	for _, resource := range resources {
		if providerMap[strings.ToLower(string(resource.Provider))] {
			filtered = append(filtered, resource)
		}
	}

	return filtered
}

// createGenerationEngine creates and configures the generation engine
func createGenerationEngine(opts *Options) (*generation.Engine, error) {
	config := generation.EngineConfig{
		MaxConcurrency:  10,
		Timeout:         opts.Timeout,
		ValidateOutput:  opts.ValidateOutput,
		DefaultFormat:   generation.IaCFormat(opts.Format),
		DefaultOrg:      generation.OrganizationPattern(opts.Organization),
		IncludeMetadata: opts.Verbose,
	}

	engine := generation.NewEngine(config)

	// Register resource mappers
	awsMapper := mappers.NewAWSMapper()
	engine.RegisterMapper(awsMapper)

	// TODO: Add Azure and GCP mappers when Phase 2 providers are ready
	// azureMapper := mappers.NewAzureMapper()
	// engine.RegisterMapper(azureMapper)
	// gcpMapper := mappers.NewGCPMapper()
	// engine.RegisterMapper(gcpMapper)

	// Register generators based on format
	switch generation.IaCFormat(opts.Format) {
	case generation.Terraform, generation.TerraformJSON:
		terraformGen := terraform.NewGenerator()
		engine.RegisterGenerator(generation.Terraform, terraformGen)
	default:
		return nil, fmt.Errorf("unsupported format: %s", opts.Format)
	}

	// TODO: Set organizer, validator, dependency analyzer
	// These would be implemented as separate components
	
	return engine, nil
}

// showGenerationPreview shows what would be generated
func showGenerationPreview(ctx context.Context, engine *generation.Engine, genOpts generation.GenerationOptions, logger *logrus.Entry) error {
	fmt.Println("🔍 IaC Generation Preview")
	fmt.Println("========================")
	
	// Get preview
	preview, err := engine.Preview(ctx, genOpts)
	if err != nil {
		return fmt.Errorf("failed to generate preview: %w", err)
	}

	// Display preview information
	fmt.Printf("Format: %s\n", genOpts.Format)
	fmt.Printf("Organization: %s\n", genOpts.Organization)
	fmt.Printf("Output Path: %s\n", genOpts.OutputPath)
	fmt.Printf("Resource Count: %d\n", preview.ResourceCount)
	fmt.Printf("Estimated Output Size: %s\n", formatBytes(preview.EstimatedSize))
	
	if len(preview.Providers) > 0 {
		fmt.Printf("\nProviders Required:\n")
		for _, provider := range preview.Providers {
			fmt.Printf("  • %s (%s) version %s\n", provider.Name, provider.Source, provider.Version)
		}
	}

	if len(preview.FileStructure) > 0 {
		fmt.Printf("\nFile Structure:\n")
		for _, file := range preview.FileStructure {
			fmt.Printf("  📄 %s (%s, %d resources, %s)\n", 
				file.Path, file.Type, file.ResourceCount, formatBytes(file.EstimatedSize))
		}
	}

	if len(preview.Variables) > 0 {
		fmt.Printf("\nVariables Required: %d\n", len(preview.Variables))
		for name, variable := range preview.Variables {
			required := ""
			if variable.Required {
				required = " (required)"
			}
			fmt.Printf("  • %s: %s%s\n", name, variable.Description, required)
		}
	}

	if len(preview.Outputs) > 0 {
		fmt.Printf("\nOutputs Generated: %d\n", len(preview.Outputs))
	}

	if len(preview.UnsupportedItems) > 0 {
		fmt.Printf("\n⚠️  Unsupported Resources: %d\n", len(preview.UnsupportedItems))
		for _, item := range preview.UnsupportedItems {
			fmt.Printf("  • %s (%s): %s\n", item.ResourceID, item.ResourceType, item.Reason)
			if item.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", item.Suggestion)
			}
		}
	}

	if len(preview.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings: %d\n", len(preview.Warnings))
		for _, warning := range preview.Warnings {
			fmt.Printf("  • %s: %s\n", warning.Type, warning.Message)
		}
	}

	fmt.Printf("\n✅ Preview complete! Use --dry-run=false to generate files.\n")
	
	return nil
}

// performGeneration performs the actual IaC generation
func performGeneration(ctx context.Context, engine *generation.Engine, genOpts generation.GenerationOptions, opts *Options, logger *logrus.Entry) error {
	fmt.Println("🔧 Generating Infrastructure as Code")
	fmt.Println("====================================")
	
	startTime := time.Now()
	
	// Perform generation
	result, err := engine.Generate(ctx, genOpts)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	duration := time.Since(startTime)

	// Display results
	fmt.Printf("✅ Generation completed in %v\n", duration)
	fmt.Printf("📊 Generation Summary:\n")
	fmt.Printf("   Files Generated: %d\n", result.Metadata.FileCount)
	fmt.Printf("   Resources Processed: %d\n", result.Metadata.ResourceCount)
	fmt.Printf("   Lines of Code: %d\n", result.Metadata.LinesGenerated)
	
	if len(result.Metadata.ProviderStats) > 0 {
		fmt.Printf("   Resources by Provider:\n")
		for provider, count := range result.Metadata.ProviderStats {
			fmt.Printf("     • %s: %d resources\n", provider, count)
		}
	}

	if genOpts.OutputPath != "" {
		fmt.Printf("   Output Directory: %s\n", genOpts.OutputPath)
	}

	// Show generated files
	if len(result.Files) > 0 {
		fmt.Printf("\n📄 Generated Files:\n")
		for _, file := range result.Files {
			size := formatBytes(file.Size)
			fmt.Printf("   • %s (%s, %s)\n", file.Path, file.Type, size)
			if file.ResourceCount > 0 {
				fmt.Printf("     Resources: %d\n", file.ResourceCount)
			}
		}
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors: %d\n", len(result.Errors))
		for _, error := range result.Errors {
			fmt.Printf("   • %s: %s\n", error.ResourceID, error.Message)
		}
	}

	// Show warnings if any
	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings: %d\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("   • %s: %s\n", warning.Type, warning.Message)
		}
	}

	// Show next steps
	fmt.Printf("\n🚀 Next Steps:\n")
	if genOpts.OutputPath != "" {
		fmt.Printf("   1. Review generated files in: %s\n", genOpts.OutputPath)
		fmt.Printf("   2. Initialize Terraform: cd %s && terraform init\n", genOpts.OutputPath)
		fmt.Printf("   3. Plan deployment: terraform plan\n")
		fmt.Printf("   4. Apply if satisfied: terraform apply\n")
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  Note: %d errors occurred during generation. Please review and fix before deployment.\n", len(result.Errors))
	}

	return nil
}

// validateOptions validates the generate command options
func validateOptions(opts *Options) error {
	if opts.InputPath == "" {
		return fmt.Errorf("input file must be specified")
	}

	if _, err := os.Stat(opts.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", opts.InputPath)
	}

	validFormats := []string{"terraform", "terraform-json", "pulumi"}
	validFormat := false
	for _, format := range validFormats {
		if opts.Format == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid format: %s (valid: %s)", 
			opts.Format, strings.Join(validFormats, ","))
	}

	validOrgPatterns := []string{"provider", "service", "region", "resource_type", "flat", "module"}
	validOrg := false
	for _, pattern := range validOrgPatterns {
		if opts.Organization == pattern {
			validOrg = true
			break
		}
	}
	if !validOrg {
		return fmt.Errorf("invalid organization pattern: %s (valid: %s)", 
			opts.Organization, strings.Join(validOrgPatterns, ","))
	}

	if len(opts.Providers) > 0 {
		validProviders := []string{"aws", "azure", "gcp"}
		for _, provider := range opts.Providers {
			valid := false
			for _, validProvider := range validProviders {
				if strings.ToLower(provider) == validProvider {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid provider: %s (valid: %s)", 
					provider, strings.Join(validProviders, ","))
			}
		}
	}

	return nil
}

// formatBytes formats byte size in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
