package generate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/pkg/generation"
)

// Options contains the generate command options
type Options struct {
	InputPath        string
	OutputPath       string
	Format           string
	OrganizeBy       string
	IncludeState     bool
	IncludeProvider  bool
	GenerateModules  bool
	ValidateOutput   bool
	ExcludeResources []string
	IncludeResources []string
	Verbose          bool
	DryRun           bool
	Timeout          time.Duration
}

// NewGenerateCommand creates the generate command
func NewGenerateCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Infrastructure as Code from discovered resources",
		Long: `Generate Infrastructure as Code templates from previously discovered
infrastructure resources. Supports multiple IaC formats including Terraform,
Pulumi, CloudFormation, and ARM templates.

Examples:
  # Generate Terraform from discovered resources
  chimera generate --input resources.json --format terraform --output ./infrastructure

  # Generate Pulumi TypeScript code
  chimera generate --input resources.json --format pulumi-typescript --output ./infra

  # Generate with modules organized by provider
  chimera generate --input resources.json --format terraform --organize-by provider --modules

  # Generate CloudFormation templates
  chimera generate --input resources.json --format cloudformation --output ./cf-templates

  # Exclude specific resource types
  chimera generate --input resources.json --format terraform --exclude s3_bucket,iam_role`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context(), opts)
		},
	}

	// Input/Output flags
	cmd.Flags().StringVarP(&opts.InputPath, "input", "i", "", 
		"Input file with discovered resources (JSON or YAML)")
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "./generated", 
		"Output directory for generated IaC files")

	// Format flags
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "terraform", 
		"IaC format (terraform,pulumi,cloudformation,arm)")
	cmd.Flags().StringVar(&opts.OrganizeBy, "organize-by", "provider", 
		"Organization method (provider,service,region,type)")

	// Generation options
	cmd.Flags().BoolVar(&opts.IncludeState, "include-state", true, 
		"Include state management files")
	cmd.Flags().BoolVar(&opts.IncludeProvider, "include-provider", true, 
		"Include provider configuration")
	cmd.Flags().BoolVar(&opts.GenerateModules, "modules", false, 
		"Generate modular structure")
	cmd.Flags().BoolVar(&opts.ValidateOutput, "validate", true, 
		"Validate generated output")

	// Filtering flags
	cmd.Flags().StringSliceVar(&opts.ExcludeResources, "exclude", []string{}, 
		"Resource types to exclude from generation")
	cmd.Flags().StringSliceVar(&opts.IncludeResources, "include", []string{}, 
		"Only include these resource types")

	// Behavior flags
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, 
		"Verbose output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, 
		"Show what would be generated without creating files")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 5*time.Minute, 
		"Generation timeout")

	// Required flags
	cmd.MarkFlagRequired("input")

	return cmd
}

// runGenerate executes the generate command
func runGenerate(ctx context.Context, opts *Options) error {
	// Setup logging
	if opts.Verbose {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logger := logrus.WithField("command", "generate")
	logger.Info("Starting IaC generation")

	// Validate options
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Load discovered resources
	resources, err := loadDiscoveredResources(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}

	logger.Infof("Loaded %d resources from %s", len(resources), opts.InputPath)

	// Parse format
	format, err := parseFormat(opts.Format)
	if err != nil {
		return fmt.Errorf("invalid format: %w", err)
	}

	// Parse organization method
	orgMethod, err := parseOrganizeBy(opts.OrganizeBy)
	if err != nil {
		return fmt.Errorf("invalid organization method: %w", err)
	}

	// Create generation options
	genOpts := generation.GenerationOptions{
		Resources:        resources,
		Format:           format,
		OutputPath:       opts.OutputPath,
		OrganizeByType:   orgMethod == generation.ModuleByResourceType,
		OrganizeByRegion: orgMethod == generation.ModuleByRegion,
		IncludeState:     opts.IncludeState,
		IncludeProvider:  opts.IncludeProvider,
		GenerateModules:  opts.GenerateModules,
		ModuleStructure:  orgMethod,
		ExcludeResources: opts.ExcludeResources,
		IncludeResources: opts.IncludeResources,
		ValidateOutput:   opts.ValidateOutput,
		Timeout:          opts.Timeout,
	}

	// Show generation plan if dry run
	if opts.DryRun {
		return showGenerationPlan(genOpts)
	}

	// Create generation engine
	engine, err := createGenerationEngine()
	if err != nil {
		return fmt.Errorf("failed to create generation engine: %w", err)
	}

	// Validate options
	if err := engine.ValidateOptions(genOpts); err != nil {
		return fmt.Errorf("invalid generation options: %w", err)
	}

	// Execute generation
	result, err := engine.Generate(ctx, genOpts)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Log summary
	logger.WithFields(logrus.Fields{
		"files_generated": result.Metadata.FileCount,
		"resources":       result.Metadata.ResourceCount,
		"duration":        result.Metadata.Duration,
		"errors":          result.Metadata.ErrorCount,
	}).Info("Generation completed")

	// Output results
	return outputGenerationResults(result, opts)
}

// validateOptions validates the generate command options
func validateOptions(opts *Options) error {
	// Check input file exists
	if _, err := os.Stat(opts.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", opts.InputPath)
	}

	// Validate format
	validFormats := []string{"terraform", "terraform-json", "pulumi", "pulumi-typescript", "pulumi-python", "cloudformation", "arm"}
	validFormat := false
	for _, format := range validFormats {
		if opts.Format == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid format: %s (valid: %s)", opts.Format, strings.Join(validFormats, ","))
	}

	// Validate organization method
	validOrgMethods := []string{"provider", "service", "region", "type", "flat"}
	validOrgMethod := false
	for _, method := range validOrgMethods {
		if opts.OrganizeBy == method {
			validOrgMethod = true
			break
		}
	}
	if !validOrgMethod {
		return fmt.Errorf("invalid organization method: %s (valid: %s)", opts.OrganizeBy, strings.Join(validOrgMethods, ","))
	}

	return nil
}

// loadDiscoveredResources loads resources from the input file
func loadDiscoveredResources(inputPath string) ([]discovery.Resource, error) {
	// This is a placeholder implementation
	// In a real implementation, this would read the JSON/YAML file
	// and unmarshal it into discovery.Resource structs
	
	logrus.Info("Loading discovered resources...")
	
	// For now, return empty slice
	// TODO: Implement actual file reading
	return []discovery.Resource{}, fmt.Errorf("resource loading not yet implemented - Phase 2 feature")
}

// parseFormat converts string format to IaCFormat enum
func parseFormat(formatStr string) (generation.IaCFormat, error) {
	switch strings.ToLower(formatStr) {
	case "terraform":
		return generation.Terraform, nil
	case "terraform-json":
		return generation.TerraformJSON, nil
	case "pulumi":
		return generation.Pulumi, nil
	case "pulumi-typescript":
		return generation.PulumiTypeScript, nil
	case "pulumi-python":
		return generation.PulumiPython, nil
	case "pulumi-go":
		return generation.PulumiGo, nil
	case "cloudformation":
		return generation.CloudFormation, nil
	case "arm":
		return generation.ARM, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", formatStr)
	}
}

// parseOrganizeBy converts string to ModuleStructure enum
func parseOrganizeBy(orgStr string) (generation.ModuleStructure, error) {
	switch strings.ToLower(orgStr) {
	case "provider":
		return generation.ModuleByProvider, nil
	case "service":
		return generation.ModuleByService, nil
	case "region":
		return generation.ModuleByRegion, nil
	case "type":
		return generation.ModuleByResourceType, nil
	case "flat":
		return generation.ModuleFlat, nil
	default:
		return "", fmt.Errorf("unsupported organization method: %s", orgStr)
	}
}

// createGenerationEngine creates and configures the generation engine
func createGenerationEngine() (generation.GenerationEngine, error) {
	// This is a placeholder implementation
	// In a real implementation, this would create the actual generation engine
	
	return nil, fmt.Errorf("generation engine not yet implemented - Phase 2 feature")
}

// showGenerationPlan shows what would be generated in a dry run
func showGenerationPlan(opts generation.GenerationOptions) error {
	fmt.Println("⚙️  Generation Plan:")
	fmt.Println("===================")
	fmt.Printf("Input Resources: %d\n", len(opts.Resources))
	fmt.Printf("Output Format: %s\n", opts.Format)
	fmt.Printf("Output Path: %s\n", opts.OutputPath)
	fmt.Printf("Organization: %s\n", opts.ModuleStructure)
	fmt.Printf("Generate Modules: %v\n", opts.GenerateModules)
	fmt.Printf("Include State: %v\n", opts.IncludeState)
	fmt.Printf("Include Provider: %v\n", opts.IncludeProvider)
	
	if len(opts.ExcludeResources) > 0 {
		fmt.Printf("Excluded Resources: %v\n", opts.ExcludeResources)
	}
	
	if len(opts.IncludeResources) > 0 {
		fmt.Printf("Included Resources: %v\n", opts.IncludeResources)
	}
	
	fmt.Printf("Validate Output: %v\n", opts.ValidateOutput)
	fmt.Printf("Timeout: %v\n", opts.Timeout)
	
	fmt.Println("\n✅ This is what would be generated.")
	fmt.Println("Remove --dry-run to execute actual generation.")
	
	return nil
}

// outputGenerationResults outputs the generation results
func outputGenerationResults(result *generation.GenerationResult, opts *Options) error {
	fmt.Printf("⚙️  Generation Summary:\n")
	fmt.Printf("======================\n")
	fmt.Printf("  Files Generated: %d\n", result.Metadata.FileCount)
	fmt.Printf("  Resources Processed: %d\n", result.Metadata.ResourceCount)
	fmt.Printf("  Lines Generated: %d\n", result.Metadata.LinesGenerated)
	fmt.Printf("  Duration: %v\n", result.Metadata.Duration)
	fmt.Printf("  Errors: %d\n", result.Metadata.ErrorCount)
	fmt.Printf("  Warnings: %d\n", result.Metadata.WarningCount)
	fmt.Printf("\n")

	// Print generated files
	if len(result.Files) > 0 {
		fmt.Println("Generated Files:")
		for _, file := range result.Files {
			fmt.Printf("  %s (%s, %d resources)\n", file.Path, file.Type, file.ResourceCount)
		}
		fmt.Printf("\n")
	}

	// Print provider statistics
	if len(result.Metadata.ProviderStats) > 0 {
		fmt.Printf("Provider Statistics:\n")
		for provider, count := range result.Metadata.ProviderStats {
			fmt.Printf("  %s: %d resources\n", provider, count)
		}
		fmt.Printf("\n")
	}

	// Print errors if any
	if len(result.Errors) > 0 {
		fmt.Printf("Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  [%s/%s] %s\n", err.Provider, err.ResourceType, err.Message)
		}
		fmt.Printf("\n")
	}

	// Print warnings if any
	if len(result.Warnings) > 0 {
		fmt.Printf("Warnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  [%s/%s] %s (%s)\n", warning.Provider, warning.ResourceType, warning.Message, warning.Type)
		}
	}

	return nil
}
