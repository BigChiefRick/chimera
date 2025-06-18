package generate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	// Input options
	InputPath        string
	InputFormat      string
	
	// Output options
	OutputPath       string
	Format           string
	OrganizeByType   bool
	OrganizeByRegion bool
	SingleFile       bool
	
	// Generation options
	IncludeState     bool
	IncludeProvider  bool
	ProviderVersion  string
	GenerateModules  bool
	ModuleStructure  string
	
	// Filtering options
	ExcludeResources []string
	IncludeResources []string
	Provider         string
	Region           string
	ResourceTypes    []string
	
	// Template options
	TemplateVariables map[string]string
	CustomTemplates   map[string]string
	
	// Behavior options
	CompactOutput    bool
	ValidateOutput   bool
	DryRun           bool
	Verbose          bool
	Force            bool
	Timeout          time.Duration
}

// NewGenerateCommand creates the generate command
func NewGenerateCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Infrastructure as Code from discovered resources",
		Long: `Generate Infrastructure as Code templates from previously discovered
infrastructure resources.

Phase 3: Full implementation with real Terraform generation, resource mapping,
dependency resolution, and multi-format support.

Examples:
  # Generate Terraform from discovered AWS resources
  chimera generate --input aws-resources.json --output ./terraform/

  # Generate organized by resource type
  chimera generate --input resources.json --output ./terraform/ --organize-by-type

  # Generate with modules
  chimera generate --input resources.json --output ./terraform/ --generate-modules

  # Preview what would be generated
  chimera generate --input resources.json --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context(), opts)
		},
	}

	// Input flags
	cmd.Flags().StringVarP(&opts.InputPath, "input", "i", "", 
		"Input file with discovered resources (required)")
	cmd.Flags().StringVar(&opts.InputFormat, "input-format", "json", 
		"Input format (json)")

	// Output flags
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "./generated", 
		"Output directory for generated files")
	cmd.Flags().StringVar(&opts.Format, "format", "terraform", 
		"Output format (terraform,pulumi,cloudformation)")
	cmd.Flags().BoolVar(&opts.OrganizeByType, "organize-by-type", false, 
		"Organize files by resource type")
	cmd.Flags().BoolVar(&opts.OrganizeByRegion, "organize-by-region", false, 
		"Organize files by region")
	cmd.Flags().BoolVar(&opts.SingleFile, "single-file", false, 
		"Generate single file with all resources")

	// Generation flags
	cmd.Flags().BoolVar(&opts.IncludeState, "include-state", false, 
		"Include state configuration")
	cmd.Flags().BoolVar(&opts.IncludeProvider, "include-provider", true, 
		"Include provider configuration")
	cmd.Flags().StringVar(&opts.ProviderVersion, "provider-version", "", 
		"Specific provider version")
	cmd.Flags().BoolVar(&opts.GenerateModules, "generate-modules", false, 
		"Generate Terraform modules")
	cmd.Flags().StringVar(&opts.ModuleStructure, "module-structure", "by_provider", 
		"Module organization (by_provider,by_service,by_region,by_resource_type)")

	// Filtering flags
	cmd.Flags().StringSliceVar(&opts.ExcludeResources, "exclude", []string{}, 
		"Resource IDs to exclude")
	cmd.Flags().StringSliceVar(&opts.IncludeResources, "include", []string{}, 
		"Resource IDs to include (if specified, only these are generated)")
	cmd.Flags().StringVar(&opts.Provider, "provider", "", 
		"Filter by cloud provider (aws,azure,gcp)")
	cmd.Flags().StringVar(&opts.Region, "region", "", 
		"Filter by region")
	cmd.Flags().StringSliceVar(&opts.ResourceTypes, "resource-type", []string{}, 
		"Specific resource types to generate")

	// Template flags
	cmd.Flags().StringToStringVar(&opts.TemplateVariables, "template-var", map[string]string{}, 
		"Template variables (key=value)")

	// Behavior flags
	cmd.Flags().BoolVar(&opts.CompactOutput, "compact", false, 
		"Generate compact output")
	cmd.Flags().BoolVar(&opts.ValidateOutput, "validate", true, 
		"Validate generated output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, 
		"Show what would be generated without creating files")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, 
		"Verbose output")
	cmd.Flags().BoolVar(&opts.Force, "force", false, 
		"Overwrite existing files")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 5*time.Minute, 
		"Generation timeout")

	// Required flags
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
	resources, err := loadResources(opts.InputPath, opts.InputFormat)
	if err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}

	logger.Infof("Loaded %d resources from %s", len(resources), opts.InputPath)

	// Filter resources if specified
	filteredResources := filterResources(resources, opts)
	logger.Infof("Filtered to %d resources", len(filteredResources))

	if len(filteredResources) == 0 {
		fmt.Println("⚠️  No resources match the specified filters")
		return nil
	}

	// Show generation plan if dry run
	if opts.DryRun {
		return showGenerationPlan(filteredResources, opts)
	}

	// Check output directory
	if err := prepareOutputDirectory(opts.OutputPath, opts.Force); err != nil {
		return fmt.Errorf("failed to prepare output directory: %w", err)
	}

	// Create generation engine
	engine := generation.NewEngine(generation.EngineConfig{
		MaxConcurrency: 10,
		Timeout:        opts.Timeout,
		ValidateOutput: opts.ValidateOutput,
	})

	// Register mappers - FIX: Remove discovery.AWS parameter
	engine.RegisterMapper(mappers.NewAWSMapper())
	// TODO: Add Azure and GCP mappers in Phase 4

	// Register generators
	engine.RegisterGenerator(generation.Terraform, terraform.NewGenerator())
	// TODO: Add Pulumi and CloudFormation generators in Phase 4

	// Convert options to generation options
	genOpts := convertToGenerationOptions(opts, filteredResources)

	// Perform generation
	result, err := engine.Generate(ctx, genOpts)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Write files to disk
	if err := writeGeneratedFiles(result.Files, opts); err != nil {
		return fmt.Errorf("failed to write files: %w", err)
	}

	// Display results
	displayResults(result, opts)

	return nil
}

// loadResources loads discovered resources from file
func loadResources(inputPath, inputFormat string) ([]discovery.Resource, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	switch strings.ToLower(inputFormat) {
	case "json":
		return loadJSONResources(data)
	default:
		return nil, fmt.Errorf("unsupported input format: %s", inputFormat)
	}
}

// loadJSONResources loads resources from JSON data
func loadJSONResources(data []byte) ([]discovery.Resource, error) {
	// Try to parse as DiscoveryResult first
	var discoveryResult discovery.DiscoveryResult
	if err := json.Unmarshal(data, &discoveryResult); err == nil && len(discoveryResult.Resources) > 0 {
		return discoveryResult.Resources, nil
	}

	// Try to parse as raw resource array
	var resources []discovery.Resource
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return resources, nil
}

// filterResources filters resources based on options
func filterResources(resources []discovery.Resource, opts *Options) []discovery.Resource {
	var filtered []discovery.Resource

	for _, resource := range resources {
		// Provider filter
		if opts.Provider != "" && string(resource.Provider) != opts.Provider {
			continue
		}

		// Region filter
		if opts.Region != "" && resource.Region != opts.Region {
			continue
		}

		// Resource type filter
		if len(opts.ResourceTypes) > 0 {
			match := false
			for _, resourceType := range opts.ResourceTypes {
				if resource.Type == resourceType {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Exclude filter
		if len(opts.ExcludeResources) > 0 {
			exclude := false
			for _, excludeID := range opts.ExcludeResources {
				if resource.ID == excludeID {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		// Include filter (if specified, only include these)
		if len(opts.IncludeResources) > 0 {
			include := false
			for _, includeID := range opts.IncludeResources {
				if resource.ID == includeID {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}

		filtered = append(filtered, resource)
	}

	return filtered
}

// showGenerationPlan displays what would be generated
func showGenerationPlan(resources []discovery.Resource, opts *Options) error {
	fmt.Printf("🎯 Generation Plan\n")
	fmt.Printf("==================\n")
	fmt.Printf("📁 Output directory: %s\n", opts.OutputPath)
	fmt.Printf("📄 Format: %s\n", opts.Format)
	fmt.Printf("🏗️  Resources to generate: %d\n", len(resources))

	// Count by provider
	providerCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, resource := range resources {
		providerCounts[string(resource.Provider)]++
		typeCounts[resource.Type]++
	}

	fmt.Printf("\n📊 Provider Breakdown:\n")
	for provider, count := range providerCounts {
		fmt.Printf("   %s: %d resources\n", provider, count)
	}

	fmt.Printf("\n🏗️  Resource Types:\n")
	for resourceType, count := range typeCounts {
		fmt.Printf("   %s: %d\n", resourceType, count)
	}

	fmt.Printf("\n📁 Files that would be generated:\n")
	if opts.SingleFile {
		fmt.Printf("   main.tf (all resources)\n")
	} else if opts.OrganizeByType {
		for resourceType := range typeCounts {
			fileName := strings.ReplaceAll(resourceType, "aws_", "") + ".tf"
			fmt.Printf("   %s\n", fileName)
		}
	} else {
		fmt.Printf("   main.tf (all resources)\n")
	}

	if opts.IncludeProvider {
		fmt.Printf("   providers.tf\n")
	}
	fmt.Printf("   variables.tf\n")
	fmt.Printf("   outputs.tf\n")

	if opts.GenerateModules {
		fmt.Printf("   modules/ (organized by %s)\n", opts.ModuleStructure)
	}

	fmt.Printf("\n✅ This is what would be generated.\n")
	fmt.Printf("Remove --dry-run to execute actual generation.\n")

	return nil
}

// validateOptions validates the generate command options
func validateOptions(opts *Options) error {
	// Check input file exists
	if _, err := os.Stat(opts.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", opts.InputPath)
	}

	// Validate format
	validFormats := []string{"terraform", "pulumi", "cloudformation"}
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

	// Validate module structure
	validStructures := []string{"by_provider", "by_service", "by_region", "by_resource_type", "flat"}
	validStructure := false
	for _, structure := range validStructures {
		if opts.ModuleStructure == structure {
			validStructure = true
			break
		}
	}
	if !validStructure {
		return fmt.Errorf("invalid module structure: %s (valid: %s)", 
			opts.ModuleStructure, strings.Join(validStructures, ","))
	}

	// Validate conflicting options
	if opts.SingleFile && opts.OrganizeByType {
		return fmt.Errorf("cannot use --single-file with --organize-by-type")
	}

	if opts.SingleFile && opts.OrganizeByRegion {
		return fmt.Errorf("cannot use --single-file with --organize-by-region")
	}

	return nil
}

// convertToGenerationOptions converts CLI options to generation options
func convertToGenerationOptions(opts *Options, resources []discovery.Resource) generation.GenerationOptions {
	// Convert format
	var format generation.IaCFormat
	switch opts.Format {
	case "terraform":
		format = generation.Terraform
	case "pulumi":
		format = generation.Pulumi
	case "cloudformation":
		format = generation.CloudFormation
	default:
		format = generation.Terraform
	}

	// Convert module structure
	var moduleStructure generation.ModuleStructure
	switch opts.ModuleStructure {
	case "by_provider":
		moduleStructure = generation.ModuleByProvider
	case "by_service":
		moduleStructure = generation.ModuleByService
	case "by_region":
		moduleStructure = generation.ModuleByRegion
	case "by_resource_type":
		moduleStructure = generation.ModuleByResourceType
	case "flat":
		moduleStructure = generation.ModuleFlat
	default:
		moduleStructure = generation.ModuleByProvider
	}

	// Convert template variables
	templateVars := make(map[string]interface{})
	for k, v := range opts.TemplateVariables {
		templateVars[k] = v
	}

	return generation.GenerationOptions{
		Resources:         resources,
		Format:            format,
		OutputPath:        opts.OutputPath,
		OrganizeByType:    opts.OrganizeByType,
		OrganizeByRegion:  opts.OrganizeByRegion,
		SingleFile:        opts.SingleFile,
		IncludeState:      opts.IncludeState,
		IncludeProvider:   opts.IncludeProvider,
		ProviderVersion:   opts.ProviderVersion,
		GenerateModules:   opts.GenerateModules,
		ModuleStructure:   moduleStructure,
		ExcludeResources:  opts.ExcludeResources,
		IncludeResources:  opts.IncludeResources,
		TemplateVariables: templateVars,
		CompactOutput:     opts.CompactOutput,
		ValidateOutput:    opts.ValidateOutput,
		Timeout:           opts.Timeout,
		Timestamp:         time.Now().Format(time.RFC3339),
	}
}

// prepareOutputDirectory prepares the output directory
func prepareOutputDirectory(outputPath string, force bool) error {
	// Check if directory exists
	if _, err := os.Stat(outputPath); err == nil {
		if !force {
			// Check if directory is empty
			entries, err := os.ReadDir(outputPath)
			if err != nil {
				return fmt.Errorf("failed to read output directory: %w", err)
			}
			
			if len(entries) > 0 {
				return fmt.Errorf("output directory %s is not empty (use --force to overwrite)", outputPath)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check output directory: %w", err)
	}

	// Create directory
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// writeGeneratedFiles writes generated files to disk
func writeGeneratedFiles(files []generation.GeneratedFile, opts *Options) error {
	for _, file := range files {
		// Create directory if it doesn't exist
		dir := filepath.Dir(file.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write file
		if err := os.WriteFile(file.Path, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}

		if opts.Verbose {
			fmt.Printf("✅ Generated: %s (%d bytes, %d resources)\n", 
				file.Path, file.Size, file.ResourceCount)
		}
	}

	return nil
}

// displayResults displays generation results
func displayResults(result *generation.GenerationResult, opts *Options) {
	fmt.Printf("🎉 Generation Complete!\n")
	fmt.Printf("=======================\n")
	fmt.Printf("📁 Output directory: %s\n", opts.OutputPath)
	fmt.Printf("📄 Files generated: %d\n", result.Metadata.FileCount)
	fmt.Printf("🏗️  Resources processed: %d\n", result.Metadata.ResourceCount)
	fmt.Printf("📝 Lines generated: %d\n", result.Metadata.LinesGenerated)
	fmt.Printf("⏱️  Duration: %v\n", result.Metadata.Duration)

	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("   %s\n", warning.Message)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("   %s\n", err.Message)
		}
	}

	fmt.Printf("\n🚀 Ready to deploy:\n")
	fmt.Printf("   cd %s\n", opts.OutputPath)
	fmt.Printf("   terraform init\n")
	fmt.Printf("   terraform plan\n")
	fmt.Printf("   terraform apply\n")
}
