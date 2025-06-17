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
  # Generate Terraform from discovery results
  chimera generate --input resources.json --format terraform --output ./terraform

  # Preview generation without creating files
  chimera generate --input resources.json --dry-run

  # Generate with specific organization
  chimera generate --input resources.json --organize-by-type --generate-modules

  # Filter and generate specific resources
  chimera generate --input resources.json --provider aws --include vpc,subnet`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context(), opts)
		},
	}

	// Input flags
	cmd.Flags().StringVarP(&opts.InputPath, "input", "i", "", 
		"Input file with discovered resources (JSON format)")
	cmd.Flags().StringVar(&opts.InputFormat, "input-format", "json", 
		"Input file format (json)")

	// Output flags
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "./generated", 
		"Output directory for generated IaC files")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "terraform", 
		"IaC format (terraform,pulumi,cloudformation)")

	// Organization flags
	cmd.Flags().BoolVar(&opts.OrganizeByType, "organize-by-type", false, 
		"Organize resources by type into separate files")
	cmd.Flags().BoolVar(&opts.OrganizeByRegion, "organize-by-region", false, 
		"Organize resources by region")
	cmd.Flags().BoolVar(&opts.SingleFile, "single-file", false, 
		"Generate all resources in a single file")

	// Generation flags
	cmd.Flags().BoolVar(&opts.IncludeState, "include-state", true, 
		"Include state management configuration")
	cmd.Flags().BoolVar(&opts.IncludeProvider, "include-provider", true, 
		"Include provider configuration")
	cmd.Flags().StringVar(&opts.ProviderVersion, "provider-version", "", 
		"Specific provider version to use")
	cmd.Flags().BoolVar(&opts.GenerateModules, "generate-modules", false, 
		"Generate Terraform modules")
	cmd.Flags().StringVar(&opts.ModuleStructure, "module-structure", "by_provider", 
		"Module organization (by_provider,by_service,by_region,by_resource_type)")

	// Filtering flags
	cmd.Flags().StringSliceVar(&opts.ExcludeResources, "exclude", []string{}, 
		"Resource types to exclude")
	cmd.Flags().StringSliceVar(&opts.IncludeResources, "include", []string{}, 
		"Resource types to include (if specified, only these will be generated)")
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

	// Register mappers
	engine.RegisterMapper(discovery.AWS, mappers.NewAWSMapper())
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
				if strings.Contains(resource.Type, resourceType) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Include filter
		if len(opts.IncludeResources) > 0 {
			match := false
			for _, include := range opts.IncludeResources {
				if strings.Contains(resource.Type, include) || strings.Contains(resource.Name, include) {
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
			excluded := false
			for _, exclude := range opts.ExcludeResources {
				if strings.Contains(resource.Type, exclude) || strings.Contains(resource.Name, exclude) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		filtered = append(filtered, resource)
	}

	return filtered
}

// convertToGenerationOptions converts command options to generation options
func convertToGenerationOptions(opts *Options, resources []discovery.Resource) generation.GenerationOptions {
	// Parse module structure
	var moduleStructure generation.ModuleStructure
	switch strings.ToLower(opts.ModuleStructure) {
	case "by_provider":
		moduleStructure = generation.ModuleByProvider
	case "by_service":
		moduleStructure = generation.ModuleByService
	case "by_region":
		moduleStructure = generation.ModuleByRegion
	case "by_resource_type":
		moduleStructure = generation.ModuleByResourceType
	default:
		moduleStructure = generation.ModuleFlat
	}

	// Parse format
	var format generation.IaCFormat
	switch strings.ToLower(opts.Format) {
	case "terraform":
		format = generation.Terraform
	case "pulumi":
		format = generation.Pulumi
	case "cloudformation":
		format = generation.CloudFormation
	default:
		format = generation.Terraform
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
	fmt.Printf("📊 Format: %s\n", result.Metadata.Format)

	// Show provider statistics
	if len(result.Metadata.ProviderStats) > 0 {
		fmt.Printf("\n📊 Provider Breakdown:\n")
		for provider, count := range result.Metadata.ProviderStats {
			fmt.Printf("   %s: %d resources\n", provider, count)
		}
	}

	// Show files created
	fmt.Printf("\n📄 Generated Files:\n")
	for _, file := range result.Files {
		fileType := ""
		switch file.Type {
		case generation.FileTypeMain:
			fileType = "main"
		case generation.FileTypeVariables:
			fileType = "variables"
		case generation.FileTypeOutputs:
			fileType = "outputs"
		case generation.FileTypeProvider:
			fileType = "provider"
		case generation.FileTypeModule:
			fileType = "module"
		default:
			fileType = "other"
		}
		
		fmt.Printf("   %s (%s, %d resources)\n", 
			file.Path, fileType, file.ResourceCount)
	}

	// Show warnings if any
	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("   %s: %s\n", warning.Type, warning.Message)
		}
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("   %s (%s): %s\n", err.ResourceType, err.Severity, err.Message)
		}
	}

	// Next steps
	fmt.Printf("\n🚀 Next Steps:\n")
	switch opts.Format {
	case "terraform":
		fmt.Printf("   1. cd %s\n", opts.OutputPath)
		fmt.Printf("   2. terraform init\n")
		fmt.Printf("   3. terraform plan\n")
		fmt.Printf("   4. terraform apply\n")
	case "pulumi":
		fmt.Printf("   1. cd %s\n", opts.OutputPath)
		fmt.Printf("   2. pulumi stack init\n")
		fmt.Printf("   3. pulumi preview\n")
		fmt.Printf("   4. pulumi up\n")
	}
}

// showGenerationPlan shows what would be generated in a dry run
func showGenerationPlan(resources []discovery.Resource, opts *Options) error {
	fmt.Printf("🔍 Generation Plan:\n")
	fmt.Printf("==================\n")
	fmt.Printf("Input: %s\n", opts.InputPath)
	fmt.Printf("Output: %s\n", opts.OutputPath)
	fmt.Printf("Format: %s\n", opts.Format)
	fmt.Printf("Resources: %d\n", len(resources))

	// Group by provider
	providerCounts := make(map[discovery.CloudProvider]int)
	typeCounts := make(map[string]int)
	
	for _, resource := range resources {
		providerCounts[resource.Provider]++
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
