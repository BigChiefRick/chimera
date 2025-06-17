package generation

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/pkg/discovery"
)

// Engine implements the GenerationEngine interface
type Engine struct {
	mappers      map[discovery.CloudProvider]ResourceMapper
	generators   map[IaCFormat]TerraformGenerator
	organizer    FileOrganizer
	validator    Validator
	analyzer     DependencyAnalyzer
	templateEng  TemplateEngine
	logger       *logrus.Logger
	config       EngineConfig
}

// EngineConfig contains configuration for the generation engine
type EngineConfig struct {
	MaxConcurrency  int           `yaml:"max_concurrency" json:"max_concurrency"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	ValidateOutput  bool          `yaml:"validate_output" json:"validate_output"`
	DefaultFormat   IaCFormat     `yaml:"default_format" json:"default_format"`
	DefaultOrg      OrganizationPattern `yaml:"default_organization" json:"default_organization"`
	IncludeMetadata bool          `yaml:"include_metadata" json:"include_metadata"`
}

// NewEngine creates a new generation engine
func NewEngine(config EngineConfig) *Engine {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 10
	}
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Minute
	}
	if config.DefaultFormat == "" {
		config.DefaultFormat = Terraform
	}
	if config.DefaultOrg == "" {
		config.DefaultOrg = OrganizeByProvider
	}

	return &Engine{
		mappers:     make(map[discovery.CloudProvider]ResourceMapper),
		generators:  make(map[IaCFormat]TerraformGenerator),
		logger:      logrus.New(),
		config:      config,
	}
}

// RegisterMapper registers a resource mapper for a specific provider
func (e *Engine) RegisterMapper(mapper ResourceMapper) {
	e.mappers[mapper.Provider()] = mapper
	e.logger.Infof("Registered resource mapper for provider: %s", mapper.Provider())
}

// RegisterGenerator registers an IaC generator for a specific format
func (e *Engine) RegisterGenerator(format IaCFormat, generator TerraformGenerator) {
	e.generators[format] = generator
	e.logger.Infof("Registered generator for format: %s", format)
}

// SetOrganizer sets the file organizer
func (e *Engine) SetOrganizer(organizer FileOrganizer) {
	e.organizer = organizer
}

// SetValidator sets the validator
func (e *Engine) SetValidator(validator Validator) {
	e.validator = validator
}

// SetDependencyAnalyzer sets the dependency analyzer
func (e *Engine) SetDependencyAnalyzer(analyzer DependencyAnalyzer) {
	e.analyzer = analyzer
}

// SetTemplateEngine sets the template engine
func (e *Engine) SetTemplateEngine(engine TemplateEngine) {
	e.templateEng = engine
}

// Generate generates IaC from discovered resources
func (e *Engine) Generate(ctx context.Context, opts GenerationOptions) (*GenerationResult, error) {
	startTime := time.Now()
	
	// Validate options
	if err := e.ValidateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid generation options: %w", err)
	}

	// Initialize result
	result := &GenerationResult{
		Files:    make([]GeneratedFile, 0),
		Errors:   make([]GenerationError, 0),
		Warnings: make([]GenerationWarning, 0),
		Metadata: GenerationMetadata{
			StartTime:     startTime,
			Format:        opts.Format,
			Organization:  opts.Organization,
			ProviderStats: make(map[discovery.CloudProvider]int),
		},
	}

	e.logger.Infof("Starting IaC generation for %d resources", len(opts.Resources))

	// Filter resources if specified
	resources := e.filterResources(opts.Resources, opts)
	e.logger.Infof("Filtered to %d resources for generation", len(resources))

	// Map resources to IaC format
	mappedResources, errors, warnings := e.mapResources(ctx, resources)
	result.Errors = append(result.Errors, errors...)
	result.Warnings = append(result.Warnings, warnings...)

	if len(mappedResources) == 0 {
		return result, fmt.Errorf("no resources could be mapped for generation")
	}

	e.logger.Infof("Successfully mapped %d resources", len(mappedResources))

	// Analyze dependencies if needed
	if e.analyzer != nil {
		if err := e.analyzeDependencies(mappedResources); err != nil {
			e.logger.Warnf("Dependency analysis failed: %v", err)
			result.Warnings = append(result.Warnings, GenerationWarning{
				Message: fmt.Sprintf("Dependency analysis failed: %v", err),
				Type:    WarningTypeBestPractice,
			})
		}
	}

	// Organize files based on pattern
	organizedFiles, err := e.organizeResources(mappedResources, opts)
	if err != nil {
		return result, fmt.Errorf("failed to organize resources: %w", err)
	}

	// Generate IaC files
	generator, exists := e.generators[opts.Format]
	if !exists {
		return result, fmt.Errorf("no generator available for format: %s", opts.Format)
	}

	files, genErrors, genWarnings := e.generateFiles(ctx, organizedFiles, generator, opts)
	result.Files = files
	result.Errors = append(result.Errors, genErrors...)
	result.Warnings = append(result.Warnings, genWarnings...)

	// Write files to disk if output path specified
	if opts.OutputPath != "" {
		if err := e.writeFiles(result.Files, opts.OutputPath); err != nil {
			return result, fmt.Errorf("failed to write files: %w", err)
		}
		e.logger.Infof("Generated files written to: %s", opts.OutputPath)
	}

	// Validate output if requested
	if opts.ValidateOutput && e.validator != nil {
		if err := e.validateOutput(result.Files, opts); err != nil {
			result.Warnings = append(result.Warnings, GenerationWarning{
				Message: fmt.Sprintf("Output validation failed: %v", err),
				Type:    WarningTypeBestPractice,
			})
		}
	}

	// Finalize metadata
	result.Metadata.EndTime = time.Now()
	result.Metadata.Duration = result.Metadata.EndTime.Sub(result.Metadata.StartTime)
	result.Metadata.ResourceCount = len(mappedResources)
	result.Metadata.FileCount = len(result.Files)
	result.Metadata.ErrorCount = len(result.Errors)
	result.Metadata.WarningCount = len(result.Warnings)

	// Calculate provider stats
	for _, resource := range mappedResources {
		result.Metadata.ProviderStats[resource.Provider]++
	}

	// Calculate total lines generated
	totalLines := 0
	for _, file := range result.Files {
		totalLines += strings.Count(file.Content, "\n")
	}
	result.Metadata.LinesGenerated = totalLines

	e.logger.Infof("Generation completed in %v", result.Metadata.Duration)
	e.logger.Infof("Generated %d files with %d lines", result.Metadata.FileCount, result.Metadata.LinesGenerated)

	return result, nil
}

// filterResources filters resources based on options
func (e *Engine) filterResources(resources []discovery.Resource, opts GenerationOptions) []discovery.Resource {
	if len(opts.IncludeResources) == 0 && len(opts.ExcludeResources) == 0 {
		return resources
	}

	var filtered []discovery.Resource
	
	for _, resource := range resources {
		// Check include list first
		if len(opts.IncludeResources) > 0 {
			included := false
			for _, include := range opts.IncludeResources {
				if strings.Contains(resource.Type, include) || strings.Contains(resource.ID, include) {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}

		// Check exclude list
		excluded := false
		for _, exclude := range opts.ExcludeResources {
			if strings.Contains(resource.Type, exclude) || strings.Contains(resource.ID, exclude) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		filtered = append(filtered, resource)
	}

	return filtered
}

// mapResources maps discovered resources to IaC resources
func (e *Engine) mapResources(ctx context.Context, resources []discovery.Resource) ([]TerraformResource, []GenerationError, []GenerationWarning) {
	var mapped []TerraformResource
	var errors []GenerationError
	var warnings []GenerationWarning

	for _, resource := range resources {
		mapper, exists := e.mappers[resource.Provider]
		if !exists {
			errors = append(errors, GenerationError{
				ResourceID:   resource.ID,
				ResourceType: resource.Type,
				Provider:     resource.Provider,
				Message:      fmt.Sprintf("no mapper available for provider: %s", resource.Provider),
				Severity:     ErrorSeverityHigh,
			})
			continue
		}

		mappedResource, err := mapper.MapResource(resource)
		if err != nil {
			errors = append(errors, GenerationError{
				ResourceID:   resource.ID,
				ResourceType: resource.Type,
				Provider:     resource.Provider,
				Message:      fmt.Sprintf("failed to map resource: %v", err),
				Error:        err,
				Severity:     ErrorSeverityMedium,
			})
			continue
		}

		// Validate mapping
		if err := mapper.ValidateMapping(resource, *mappedResource); err != nil {
			warnings = append(warnings, GenerationWarning{
				ResourceID:   resource.ID,
				ResourceType: resource.Type,
				Provider:     resource.Provider,
				Message:      fmt.Sprintf("mapping validation warning: %v", err),
				Type:         WarningTypeBestPractice,
			})
		}

		mapped = append(mapped, *mappedResource)
	}

	return mapped, errors, warnings
}

// analyzeDependencies analyzes resource dependencies
func (e *Engine) analyzeDependencies(resources []TerraformResource) error {
	if e.analyzer == nil {
		return nil
	}

	// Convert TerraformResource back to discovery.Resource for analysis
	discoveryResources := make([]discovery.Resource, len(resources))
	for i, resource := range resources {
		discoveryResources[i] = discovery.Resource{
			ID:       resource.SourceInfo.OriginalID,
			Type:     resource.SourceInfo.OriginalType,
			Provider: resource.SourceInfo.OriginalProvider,
			Region:   resource.SourceInfo.OriginalRegion,
			Metadata: resource.SourceInfo.Metadata,
			Tags:     resource.SourceInfo.Tags,
		}
	}

	dependencies, err := e.analyzer.AnalyzeDependencies(discoveryResources)
	if err != nil {
		return fmt.Errorf("dependency analysis failed: %w", err)
	}

	// Update resources with dependency information
	for i := range resources {
		resourceID := resources[i].SourceInfo.OriginalID
		if deps, exists := dependencies[resourceID]; exists {
			resources[i].Dependencies = deps
		}
	}

	return nil
}

// organizeResources organizes resources into file structure
func (e *Engine) organizeResources(resources []TerraformResource, opts GenerationOptions) (map[string][]TerraformResource, error) {
	if e.organizer == nil {
		// Default flat organization
		return map[string][]TerraformResource{
			"main.tf": resources,
		}, nil
	}

	organization := opts.Organization
	if organization == "" {
		organization = e.config.DefaultOrg
	}

	return e.organizer.OrganizeFiles(resources, organization)
}

// generateFiles generates the actual IaC files
func (e *Engine) generateFiles(ctx context.Context, organizedFiles map[string][]TerraformResource, generator TerraformGenerator, opts GenerationOptions) ([]GeneratedFile, []GenerationError, []GenerationWarning) {
	var files []GeneratedFile
	var errors []GenerationError
	var warnings []GenerationWarning

	// Collect all providers needed
	providerConfigs := e.collectProviderConfigs(organizedFiles, opts)

	// Generate provider configuration files
	if opts.IncludeProvider && len(providerConfigs) > 0 {
		// Generate versions.tf
		versionsContent, err := generator.GenerateVersions(providerConfigs)
		if err != nil {
			errors = append(errors, GenerationError{
				Message:  fmt.Sprintf("failed to generate versions.tf: %v", err),
				Severity: ErrorSeverityMedium,
			})
		} else {
			files = append(files, GeneratedFile{
				Path:    "versions.tf",
				Content: versionsContent,
				Type:    FileTypeVersions,
				Format:  opts.Format,
				Size:    int64(len(versionsContent)),
			})
		}

		// Generate providers.tf
		for _, config := range providerConfigs {
			providerContent, err := generator.GenerateProvider(config)
			if err != nil {
				errors = append(errors, GenerationError{
					Message:  fmt.Sprintf("failed to generate provider %s: %v", config.Name, err),
					Severity: ErrorSeverityMedium,
				})
				continue
			}

			files = append(files, GeneratedFile{
				Path:    fmt.Sprintf("provider_%s.tf", config.Name),
				Content: providerContent,
				Type:    FileTypeProvider,
				Format:  opts.Format,
				Size:    int64(len(providerContent)),
			})
		}
	}

	// Generate main resource files
	for filePath, resources := range organizedFiles {
		content, err := e.generateResourceFile(resources, generator)
		if err != nil {
			errors = append(errors, GenerationError{
				File:     filePath,
				Message:  fmt.Sprintf("failed to generate file %s: %v", filePath, err),
				Severity: ErrorSeverityHigh,
			})
			continue
		}

		files = append(files, GeneratedFile{
			Path:          filePath,
			Content:       content,
			Type:          FileTypeMain,
			Format:        opts.Format,
			Size:          int64(len(content)),
			ResourceCount: len(resources),
			Checksum:      e.calculateChecksum(content),
		})
	}

	// Generate variables.tf if needed
	variables := e.collectVariables(organizedFiles)
	if len(variables) > 0 {
		variablesContent, err := generator.GenerateVariables(variables)
		if err != nil {
			warnings = append(warnings, GenerationWarning{
				Message: fmt.Sprintf("failed to generate variables.tf: %v", err),
				Type:    WarningTypeBestPractice,
			})
		} else {
			files = append(files, GeneratedFile{
				Path:    "variables.tf",
				Content: variablesContent,
				Type:    FileTypeVariables,
				Format:  opts.Format,
				Size:    int64(len(variablesContent)),
			})
		}
	}

	// Generate outputs.tf if needed
	outputs := e.collectOutputs(organizedFiles)
	if len(outputs) > 0 {
		outputsContent, err := generator.GenerateOutputs(outputs)
		if err != nil {
			warnings = append(warnings, GenerationWarning{
				Message: fmt.Sprintf("failed to generate outputs.tf: %v", err),
				Type:    WarningTypeBestPractice,
			})
		} else {
			files = append(files, GeneratedFile{
				Path:    "outputs.tf",
				Content: outputsContent,
				Type:    FileTypeOutputs,
				Format:  opts.Format,
				Size:    int64(len(outputsContent)),
			})
		}
	}

	return files, errors, warnings
}

// generateResourceFile generates content for a single resource file
func (e *Engine) generateResourceFile(resources []TerraformResource, generator TerraformGenerator) (string, error) {
	var content strings.Builder

	content.WriteString("# Generated by Chimera\n")
	content.WriteString(fmt.Sprintf("# Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	for _, resource := range resources {
		resourceContent, err := generator.GenerateResource(resource)
		if err != nil {
			return "", fmt.Errorf("failed to generate resource %s: %w", resource.Name, err)
		}

		content.WriteString(resourceContent)
		content.WriteString("\n\n")
	}

	return content.String(), nil
}

// collectProviderConfigs collects all unique provider configurations needed
func (e *Engine) collectProviderConfigs(organizedFiles map[string][]TerraformResource, opts GenerationOptions) []ProviderConfig {
	providerMap := make(map[discovery.CloudProvider]ProviderConfig)

	for _, resources := range organizedFiles {
		for _, resource := range resources {
			if mapper, exists := e.mappers[resource.Provider]; exists {
				config, err := mapper.GetProviderConfig([]discovery.Resource{
					{
						Provider: resource.Provider,
						Region:   resource.SourceInfo.OriginalRegion,
					},
				})
				if err == nil {
					providerMap[resource.Provider] = *config
				}
			}
		}
	}

	var configs []ProviderConfig
	for _, config := range providerMap {
		configs = append(configs, config)
	}

	return configs
}

// collectVariables collects all variables from resources
func (e *Engine) collectVariables(organizedFiles map[string][]TerraformResource) map[string]Variable {
	variables := make(map[string]Variable)

	for _, resources := range organizedFiles {
		for _, resource := range resources {
			for name, variable := range resource.Variables {
				if existing, exists := variables[name]; exists {
					// Merge variable definitions, preferring required ones
					if variable.Required && !existing.Required {
						variables[name] = variable
					}
				} else {
					variables[name] = variable
				}
			}
		}
	}

	return variables
}

// collectOutputs collects all outputs from resources
func (e *Engine) collectOutputs(organizedFiles map[string][]TerraformResource) map[string]Output {
	outputs := make(map[string]Output)

	for _, resources := range organizedFiles {
		for _, resource := range resources {
			for name, outputValue := range resource.Outputs {
				outputs[name] = Output{
					Name:        name,
					Value:       outputValue,
					Description: fmt.Sprintf("Output for %s %s", resource.Type, resource.Name),
				}
			}
		}
	}

	return outputs
}

// writeFiles writes generated files to disk
func (e *Engine) writeFiles(files []GeneratedFile, outputPath string) error {
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, file := range files {
		filePath := filepath.Join(outputPath, file.Path)
		
		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}

		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}

	return nil
}

// validateOutput validates generated files
func (e *Engine) validateOutput(files []GeneratedFile, opts GenerationOptions) error {
	if e.validator == nil {
		return nil
	}

	for _, file := range files {
		if err := e.validator.ValidateSyntax(file.Content, opts.Format); err != nil {
			return fmt.Errorf("validation failed for %s: %w", file.Path, err)
		}
	}

	return nil
}

// calculateChecksum calculates SHA256 checksum for content
func (e *Engine) calculateChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// ListFormats returns supported IaC formats
func (e *Engine) ListFormats() []IaCFormat {
	var formats []IaCFormat
	for format := range e.generators {
		formats = append(formats, format)
	}
	return formats
}

// ValidateOptions validates generation options
func (e *Engine) ValidateOptions(opts GenerationOptions) error {
	if len(opts.Resources) == 0 {
		return fmt.Errorf("no resources provided for generation")
	}

	if opts.Format == "" {
		return fmt.Errorf("output format must be specified")
	}

	if _, exists := e.generators[opts.Format]; !exists {
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}

	if opts.Organization == "" {
		opts.Organization = e.config.DefaultOrg
	}

	return nil
}

// GetFormatCapabilities returns capabilities for a specific format
func (e *Engine) GetFormatCapabilities(format IaCFormat) FormatCapabilities {
	capabilities := FormatCapabilities{
		Format:             format,
		SupportedProviders: make([]discovery.CloudProvider, 0),
		SupportedResources: make(map[string][]string),
		OrganizationPatterns: []OrganizationPattern{
			OrganizeByProvider,
			OrganizeByService,
			OrganizeByRegion,
			OrganizeByResourceType,
			OrganizeFlat,
		},
	}

	// Collect supported providers and resources
	for provider, mapper := range e.mappers {
		capabilities.SupportedProviders = append(capabilities.SupportedProviders, provider)
		capabilities.SupportedResources[string(provider)] = mapper.GetSupportedTypes()
	}

	// Set format-specific capabilities
	switch format {
	case Terraform, TerraformJSON:
		capabilities.SupportsModules = true
		capabilities.SupportsState = true
		capabilities.SupportsVariables = true
		capabilities.SupportsOutputs = true
		capabilities.SupportsValidation = true
		capabilities.SupportsImports = true
	default:
		// Other formats have limited support for now
		capabilities.SupportsModules = false
		capabilities.SupportsState = false
		capabilities.SupportsVariables = false
		capabilities.SupportsOutputs = false
		capabilities.SupportsValidation = false
		capabilities.SupportsImports = false
	}

	return capabilities
}

// Preview generates a preview of what would be generated
func (e *Engine) Preview(ctx context.Context, opts GenerationOptions) (*GenerationPreview, error) {
	preview := &GenerationPreview{
		FileStructure:    make([]PreviewFile, 0),
		UnsupportedItems: make([]UnsupportedResource, 0),
		Warnings:         make([]GenerationWarning, 0),
		Providers:        make([]ProviderConfig, 0),
		Variables:        make(map[string]Variable),
		Outputs:          make(map[string]Output),
	}

	// Filter resources
	resources := e.filterResources(opts.Resources, opts)
	preview.ResourceCount = len(resources)

	// Check for unsupported resources
	for _, resource := range resources {
		if _, exists := e.mappers[resource.Provider]; !exists {
			preview.UnsupportedItems = append(preview.UnsupportedItems, UnsupportedResource{
				ResourceID:   resource.ID,
				ResourceType: resource.Type,
				Provider:     resource.Provider,
				Reason:       fmt.Sprintf("No mapper available for provider %s", resource.Provider),
				Suggestion:   "Check if the provider is supported and properly configured",
			})
		}
	}

	// Estimate file structure
	if e.organizer != nil {
		organization := opts.Organization
		if organization == "" {
			organization = e.config.DefaultOrg
		}

		// Simple estimation - in real implementation, this would be more detailed
		preview.FileStructure = append(preview.FileStructure, PreviewFile{
			Path:          "main.tf",
			Type:          FileTypeMain,
			ResourceCount: len(resources),
			EstimatedSize: int64(len(resources) * 200), // Rough estimation
		})

		if opts.IncludeProvider {
			preview.FileStructure = append(preview.FileStructure, PreviewFile{
				Path:          "versions.tf",
				Type:          FileTypeVersions,
				EstimatedSize: 500,
			})
		}
	}

	// Estimate total size
	totalSize := int64(0)
	for _, file := range preview.FileStructure {
		totalSize += file.EstimatedSize
	}
	preview.EstimatedSize = totalSize

	return preview, nil
}
