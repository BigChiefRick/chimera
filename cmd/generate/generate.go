package generate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
)

// Options contains the generate command options
type Options struct {
	InputPath        string
	OutputPath       string
	Format           string
	OrganizeBy       string
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
infrastructure resources.

Phase 1: Basic framework and interface design
Phase 2: Full implementation with Terraform, Pulumi, CloudFormation support

Examples:
  # Show generation plan (dry-run)
  chimera generate --input resources.json --format terraform --dry-run
  
  # Generate Terraform (Phase 2)
  chimera generate --input resources.json --format terraform --output ./infrastructure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context(), opts)
		},
	}

	// Input/Output flags
	cmd.Flags().StringVarP(&opts.InputPath, "input", "i", "", 
		"Input file with discovered resources")
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "./generated", 
		"Output directory for generated IaC files")

	// Format flags
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "terraform", 
		"IaC format (terraform,pulumi,cloudformation)")
	cmd.Flags().StringVar(&opts.OrganizeBy, "organize-by", "provider", 
		"Organization method (provider,service,region)")

	// Behavior flags
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, 
		"Verbose output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, 
		"Show what would be generated without creating files")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 5*time.Minute, 
		"Generation timeout")

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

	// Show generation plan if dry run
	if opts.DryRun {
		return showGenerationPlan(opts)
	}

	// Phase 1 - Show framework is ready
	fmt.Println("⚙️  Generation Framework (Phase 1 - Interface Complete)")
	fmt.Println("======================================================")
	fmt.Printf("Input: %s\n", opts.InputPath)
	fmt.Printf("Output: %s\n", opts.OutputPath)
	fmt.Printf("Format: %s\n", opts.Format)
	fmt.Printf("Organization: %s\n", opts.OrganizeBy)
	fmt.Println("")
	fmt.Println("✅ Generation interfaces defined")
	fmt.Println("✅ CLI framework complete")
	fmt.Println("⚠️  Note: Full implementation coming in Phase 2")
	fmt.Println("")
	fmt.Println("Phase 2 will support:")
	fmt.Println("  • Terraform generation")
	fmt.Println("  • Pulumi generation")
	fmt.Println("  • CloudFormation generation")
	fmt.Println("  • Multi-provider modules")

	return nil
}

// validateOptions validates the generate command options
func validateOptions(opts *Options) error {
	// For Phase 1, basic validation
	validFormats := []string{"terraform", "pulumi", "cloudformation", "arm"}
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

	validOrgMethods := []string{"provider", "service", "region", "type"}
	validOrgMethod := false
	for _, method := range validOrgMethods {
		if opts.OrganizeBy == method {
			validOrgMethod = true
			break
		}
	}
	if !validOrgMethod {
		return fmt.Errorf("invalid organization method: %s (valid: %s)", 
			opts.OrganizeBy, strings.Join(validOrgMethods, ","))
	}

	return nil
}

// showGenerationPlan shows what would be generated in a dry run
func showGenerationPlan(opts *Options) error {
	fmt.Println("⚙️  Generation Plan:")
	fmt.Println("===================")
	fmt.Printf("Input File: %s\n", opts.InputPath)
	fmt.Printf("Output Directory: %s\n", opts.OutputPath)
	fmt.Printf("IaC Format: %s\n", opts.Format)
	fmt.Printf("Organization: %s\n", opts.OrganizeBy)
	fmt.Printf("Timeout: %v\n", opts.Timeout)
	
	fmt.Println("\n📋 What would be generated (Phase 2):")
	switch opts.Format {
	case "terraform":
		fmt.Println("  • main.tf - Resource definitions")
		fmt.Println("  • variables.tf - Input variables")
		fmt.Println("  • outputs.tf - Output values")
		fmt.Println("  • provider.tf - Provider configuration")
	case "pulumi":
		fmt.Println("  • __main__.py - Pulumi program")
		fmt.Println("  • requirements.txt - Dependencies")
		fmt.Println("  • Pulumi.yaml - Project configuration")
	case "cloudformation":
		fmt.Println("  • template.yaml - CloudFormation template")
		fmt.Println("  • parameters.json - Template parameters")
	}
	
	fmt.Println("\n✅ This is what would be generated in Phase 2.")
	fmt.Println("Remove --dry-run to execute actual generation.")
	
	return nil
}
