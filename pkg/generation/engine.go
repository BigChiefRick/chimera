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
func (e *Engine) SetTemplateEngine(templateEng TemplateEngine) {
	e.templateEng = templateEng
}

// Generate generates IaC from discovered resources
func (e *Engine) Generate(ctx context.Context, opts GenerationOptions) (*GenerationResult, error) {
	// Start timing
	startTime := time.Now()

	// Validate options
	if err := e.ValidateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Apply defaults
	if opts.Format == "" {
		opts.Format = e.config.DefaultFormat
	}
	if opts.Organization == "" {
		opts.Organization = e.config.DefaultOrg
	}

	// Get generator
	generator, exists := e.generators[opts.Format]
	if !exists {
		return nil, fmt.Errorf("no generator available for format: %s", opts.Format)
	}

	// Initialize result
	result := &GenerationResult{
		Files: make([]GeneratedFile, 0),
		Metadata: GenerationMetadata{
			StartTime:     startTime,
			Format:        opts.Format,
			Organization:  opts.Organization,
			ProviderStats: make(map[discovery.CloudProvider]int),
		},
		Errors:   make([]GenerationError, 0),
		Warnings: make([]GenerationWarning, 0),
	}

	// Filter resources
	resources := e.filterResources(opts.Resources, opts)
	if len(resources) == 0 {
		return result, fmt.Errorf("no resources to generate after filtering")
	}

	// Map resources to Terraform format
	mappedResources, mappingErrors, mappingWarnings := e.mapResources(ctx, resources, opts)
	result.Errors = append(result.Errors, mappingErrors...)
	result.Warnings = append(result.Warnings, mappingWarnings...)

	if len(mappedResources) == 0 {
		return result, fmt.Errorf("no resources could be mapped for generation")
	}

	// Analyze dependencies if analyzer is available
	if e.analyzer != nil {
		if err := e.analyzeDependencies(ctx, mappedResources); err != nil {
			result.Warnings = append(result.Warnings, GenerationWarning{
				Message: fmt.Sprintf("dependency analysis failed: %v", err),
				Type:    WarningTypeBestPractice,
			})
		}
	}

	// Organize resources into files
	organizedFiles, err := e.organizeResources(mappedResources, opts)
	if err != nil {
		return result, fmt.Errorf("failed to organize resources: %w", err)
	}

	// Generate files
	genFiles, genErrors, genWarnings := e.generateFiles(ctx, organizedFiles, generator, opts)
	result.Files = append(result.Files, genFiles...)
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
		result.Metadata.ProviderStats[resource.OriginalResource.Provider]++
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

// mapResources maps discovery resources to Terraform resources
func (e *Engine) mapResources(ctx context.Context, resources []discovery.Resource, opts GenerationOptions) ([]MappedResource, []GenerationError, []GenerationWarning) {
	var mapped []MappedResource
	var errors []GenerationError
	var warnings []GenerationWarning

	for _, resource := range resources {
		mapper, exists := e.mappers[resource.Provider]
		if !exists {
			errors = append(errors, GenerationError{
				ResourceID:   resource.ID,
				ResourceType: resource.Type,
				Provider:     resource.Provider,
				Message:      fmt.Sprintf("No mapper available for provider %s", resource.Provider),
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
				Message:      fmt.Sprintf("Failed to map resource: %v", err),
				Severity:     ErrorSeverityMedium,
			})
			continue
		}

		mapped = append(mapped, *mappedResource)
	}

	return mapped, errors, warnings
}

// analyzeDependencies analyzes dependencies between resources
func (e *Engine) analyzeDependencies(ctx context.Context, resources []MappedResource) error {
	// Convert MappedResource back to discovery.Resource for analysis
	discoveryResources := make([]discovery.Resource, len(resources))
	for i, resource := range resources {
		discoveryResources[i] = resource.OriginalResource
	}

	dependencies, err := e.analyzer.AnalyzeDependencies(discoveryResources)
	if err != nil {
		return fmt.Errorf("dependency analysis failed: %w", err)
	}

	// Update resources with dependency information
	for i := range resources {
		resourceID := resources[i].OriginalResource.ID
		if deps, exists := dependencies[resourceID]; exists {
			resources[i].Dependencies = deps
		}
	}

	return nil
}

// organizeResources organizes resources into file structure
func (e *Engine) organizeResources(resources []MappedResource, opts GenerationOptions) (map[string][]MappedResource, error) {
	if e.organizer == nil {
		// Default flat organization
		return map[string][]MappedResource{
			"main.tf": resources,
		}, nil
	}

	// Convert MappedResource to TerraformResource for organizer interface
	terraformResources := make([]TerraformResource, len(resources))
	for i, mapped := range resources {
		terraformResources[i] = TerraformResource{
			Type:         mapped.ResourceType,
			Name:         mapped.ResourceName,
			Provider:     mapped.OriginalResource.Provider,
			Config:       mapped.Configuration,
			Dependencies: mapped.Dependencies,
			Outputs:      make(map[string]string), // Convert from map[string]Output
			Variables:    mapped.Variables,
			SourceInfo: SourceInfo{
				OriginalID:       mapped.OriginalResource.ID,
				OriginalType:     mapped.OriginalResource.Type,
				OriginalProvider: mapped.OriginalResource.Provider,
				OriginalRegion:   mapped.OriginalResource.Region,
				DiscoveredAt:     time.Now(), // Use current time for now
				Metadata:         mapped.OriginalResource.Metadata,
				Tags:             mapped.OriginalResource.Tags,
			},
		}
		// Convert outputs
		for name, output := range mapped.Outputs {
			terraformResources[i].Outputs[name] = output.Value
		}
	}

	organization := opts.Organization
	if organization == "" {
		organization = e.config.DefaultOrg
	}

	organizedTerraform, err := e.organizer.OrganizeFiles(terraformResources, organization)
	if err != nil {
		return nil, err
	}

	// Convert back to MappedResource
	result := make(map[string][]MappedResource)
	for path, tfResources := range organizedTerraform {
		mappedResources := make([]MappedResource, len(tfResources))
		for i, tfRes := range tfResources {
			// Find the original mapped resource
			for _, orig := range resources {
				if orig.ResourceName == tfRes.Name && orig.ResourceType == tfRes.Type {
					mappedResources[i] = orig
					break
				}
			}
		}
		result[path] = mappedResources
	}

	return result, nil
}

// generateFiles generates the actual IaC files
func (e *Engine) generateFiles(ctx context.Context, organizedFiles map[string][]MappedResource, generator TerraformGenerator, opts GenerationOptions) ([]GeneratedFile, []GenerationError, []GenerationWarning) {
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
func (e *Engine) generateResourceFile(resources []MappedResource, generator TerraformGenerator) (string, error) {
	var content strings.Builder

	content.WriteString("# Generated by Chimera\n")
	content.WriteString(fmt.Sprintf("# Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	for _, resource := range resources {
		resourceContent, err := generator.GenerateResourceHCL(resource)
		if err != nil {
			return "", fmt.Errorf("failed to generate resource %s: %w", resource.ResourceName, err)
		}

		content.WriteString(resourceContent)
		content.WriteString("\n\n")
	}

	return content.String(), nil
}

// collectProviderConfigs collects all unique provider configurations needed
func (e *Engine) collectProviderConfigs(organizedFiles map[string][]MappedResource, opts GenerationOptions) []ProviderConfig {
	providerMap := make(map[discovery.CloudProvider]ProviderConfig)

	for _, resources := range organizedFiles {
		for _, resource := range resources {
			if mapper, exists := e.mappers[resource.OriginalResource.Provider]; exists {
				config, err := mapper.GetProviderConfig([]discovery.Resource{resource.OriginalResource})
				if err == nil {
					providerMap[resource.OriginalResource.Provider] = *config
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
func (e *Engine) collectVariables(organizedFiles map[string][]MappedResource) map[string]Variable {
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
func (e *Engine) collectOutputs(organizedFiles map[string][]MappedResource) map[string]Output {
	outputs := make(map[string]Output)

	for _, resources := range organizedFiles {
		for _, resource := range resources {
			for name, output := range resource.Outputs {
				outputs[name] = output
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
