package generation

import (
	"context"
	"time"

	"github.com/BigChiefRick/chimera/pkg/discovery"
)

// IaCFormat represents the different Infrastructure as Code formats supported
type IaCFormat string

const (
	Terraform        IaCFormat = "terraform"
	TerraformJSON    IaCFormat = "terraform-json"
	CloudFormation   IaCFormat = "cloudformation"
	ARM              IaCFormat = "arm"
	Pulumi           IaCFormat = "pulumi"
	PulumiTypeScript IaCFormat = "pulumi-typescript"
	PulumiPython     IaCFormat = "pulumi-python"
	PulumiGo         IaCFormat = "pulumi-go"
	PulumiCSharp     IaCFormat = "pulumi-csharp"
	PulumiJava       IaCFormat = "pulumi-java"
	CDK              IaCFormat = "cdk"
	CDKTypeScript    IaCFormat = "cdk-typescript"
	CDKPython        IaCFormat = "cdk-python"
	CDKJava          IaCFormat = "cdk-java"
	CDKCSharp        IaCFormat = "cdk-csharp"
)

// ModuleStructure defines how to structure generated modules
type ModuleStructure string

const (
	ModuleFlat           ModuleStructure = "flat"
	ModuleByProvider     ModuleStructure = "by_provider"
	ModuleByService      ModuleStructure = "by_service"
	ModuleByRegion       ModuleStructure = "by_region"
	ModuleByResourceType ModuleStructure = "by_resource_type"
)

// GenerationOptions contains configuration for IaC generation
type GenerationOptions struct {
	// Input
	Resources []discovery.Resource `json:"resources"`
	
	// Output format
	Format IaCFormat `json:"format"`
	
	// Output configuration
	OutputPath       string `json:"output_path"`
	OrganizeByType   bool   `json:"organize_by_type"`
	OrganizeByRegion bool   `json:"organize_by_region"`
	SingleFile       bool   `json:"single_file"`
	
	// Generation settings
	IncludeState     bool              `json:"include_state"`
	IncludeProvider  bool              `json:"include_provider"`
	ProviderVersion  string            `json:"provider_version,omitempty"`
	GenerateModules  bool              `json:"generate_modules"`
	ModuleStructure  ModuleStructure   `json:"module_structure"`
	
	// Filtering and transformation
	ExcludeResources []string          `json:"exclude_resources,omitempty"`
	IncludeResources []string          `json:"include_resources,omitempty"`
	ResourceMapping  map[string]string `json:"resource_mapping,omitempty"`
	
	// Template customization
	TemplateVariables map[string]interface{} `json:"template_variables,omitempty"`
	CustomTemplates   map[string]string      `json:"custom_templates,omitempty"`
	
	// Advanced options
	CompactOutput    bool          `json:"compact_output"`
	ValidateOutput   bool          `json:"validate_output"`
	GenerateImports  bool          `json:"generate_imports"`
	Timeout          time.Duration `json:"timeout"`
	Timestamp        string        `json:"timestamp"`
}

// GenerationResult contains the results of IaC generation
type GenerationResult struct {
	Files     []GeneratedFile    `json:"files"`
	Metadata  GenerationMetadata `json:"metadata"`
	Errors    []GenerationError  `json:"errors,omitempty"`
	Warnings  []GenerationWarning `json:"warnings,omitempty"`
}

// GeneratedFile represents a generated IaC file
type GeneratedFile struct {
	Path         string    `json:"path"`
	Content      string    `json:"content"`
	Type         FileType  `json:"type"`
	Format       IaCFormat `json:"format"`
	Size         int64     `json:"size"`
	ResourceCount int      `json:"resource_count"`
	Checksum     string    `json:"checksum,omitempty"`
}

// FileType represents the type of generated file
type FileType string

const (
	FileTypeMain         FileType = "main"
	FileTypeVariables    FileType = "variables"
	FileTypeOutputs      FileType = "outputs"
	FileTypeProvider     FileType = "provider"
	FileTypeVersions     FileType = "versions"
	FileTypeState        FileType = "state"
	FileTypeModule       FileType = "module"
	FileTypeData         FileType = "data"
	FileTypeImports      FileType = "imports"
	FileTypeTerraformRC  FileType = "terraformrc"
)

// GenerationMetadata contains metadata about the generation operation
type GenerationMetadata struct {
	StartTime      time.Time                      `json:"start_time"`
	EndTime        time.Time                      `json:"end_time"`
	Duration       time.Duration                  `json:"duration"`
	ResourceCount  int                            `json:"resource_count"`
	FileCount      int                            `json:"file_count"`
	LinesGenerated int                            `json:"lines_generated"`
	Format         IaCFormat                      `json:"format"`
	ProviderStats  map[discovery.CloudProvider]int `json:"provider_stats"`
	ErrorCount     int                            `json:"error_count"`
	WarningCount   int                            `json:"warning_count"`
}

// GenerationError represents an error during generation
type GenerationError struct {
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Provider     discovery.CloudProvider `json:"provider"`
	Message      string                 `json:"message"`
	Error        error                  `json:"-"`
	Severity     string                 `json:"severity"`
	Timestamp    time.Time              `json:"timestamp"`
}

// GenerationWarning represents a warning during generation
type GenerationWarning struct {
	ResourceID string      `json:"resource_id,omitempty"`
	Message    string      `json:"message"`
	Type       WarningType `json:"type"`
	Timestamp  time.Time   `json:"timestamp"`
}

// WarningType represents the type of warning
type WarningType string

const (
	WarningTypeUnsupported   WarningType = "unsupported"
	WarningTypeBestPractice  WarningType = "best_practice"
	WarningTypeConfiguration WarningType = "configuration"
	WarningTypeSecurity      WarningType = "security"
	WarningTypePerformance   WarningType = "performance"
)

// MappedResource represents a discovered resource mapped to IaC format
type MappedResource struct {
	OriginalResource discovery.Resource         `json:"original_resource"`
	ResourceType     string                     `json:"resource_type"`     // e.g., "aws_vpc"
	ResourceName     string                     `json:"resource_name"`     // Terraform resource name
	Configuration    map[string]interface{}     `json:"configuration"`     // Resource configuration
	Dependencies     []string                   `json:"dependencies"`      // Resource dependencies
	Variables        map[string]Variable        `json:"variables"`         // Associated variables
	Outputs          map[string]Output          `json:"outputs"`           // Associated outputs
	Metadata         map[string]interface{}     `json:"metadata"`          // Additional metadata
}

// Variable represents a Terraform variable
type Variable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`        // e.g., "string", "number", "bool"
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Validation  []ValidationRule `json:"validation,omitempty"`
	Sensitive   bool        `json:"sensitive"`
	Required    bool        `json:"required"`
}

// ValidationRule represents a variable validation rule
type ValidationRule struct {
	Condition    string `json:"condition"`
	ErrorMessage string `json:"error_message"`
}

// Output represents a Terraform output
type Output struct {
	Name        string      `json:"name"`
	Value       string      `json:"value"`       // Terraform expression
	Description string      `json:"description"`
	Sensitive   bool        `json:"sensitive"`
	Type        string      `json:"type,omitempty"`
}

// Generator defines the interface for IaC generators
type Generator interface {
	// Generate generates IaC files from mapped resources
	Generate(resources []MappedResource, opts GenerationOptions) ([]GeneratedFile, error)
	
	// ValidateOutput validates generated output
	ValidateOutput(files []GeneratedFile) error
	
	// GetSupportedFormats returns supported IaC formats
	GetSupportedFormats() []IaCFormat
}

// ResourceMapper defines the interface for mapping discovered resources to IaC resources
type ResourceMapper interface {
	// MapResources maps discovered resources to IaC representations
	MapResources(resources []discovery.Resource, opts GenerationOptions) ([]MappedResource, error)
	
	// GetSupportedResourceTypes returns supported resource types for this mapper
	GetSupportedResourceTypes() []string
	
	// GetProvider returns the cloud provider this mapper handles
	GetProvider() discovery.CloudProvider
}

// TemplateEngine defines the interface for template processing
type TemplateEngine interface {
	// ProcessTemplate processes a template with the given data
	ProcessTemplate(template string, data interface{}) (string, error)
	
	// LoadTemplate loads a template from file
	LoadTemplate(path string) (string, error)
	
	// RegisterFunction registers a custom template function
	RegisterFunction(name string, fn interface{}) error
	
	// ValidateTemplate validates template syntax
	ValidateTemplate(template string) error
}

// FileOrganizer defines the interface for organizing generated files
type FileOrganizer interface {
	// OrganizeFiles organizes files based on the specified pattern
	OrganizeFiles(files []GeneratedFile, pattern ModuleStructure) ([]GeneratedFile, error)
	
	// GroupByProvider groups files by cloud provider
	GroupByProvider(files []GeneratedFile) map[discovery.CloudProvider][]GeneratedFile
	
	// GroupByResourceType groups files by resource type
	GroupByResourceType(files []GeneratedFile) map[string][]GeneratedFile
	
	// GroupByRegion groups files by region
	GroupByRegion(files []GeneratedFile) map[string][]GeneratedFile
}

// Validator defines the interface for validating generated IaC
type Validator interface {
	// ValidateFile validates a single generated file
	ValidateFile(path string, content string, format IaCFormat) error
	
	// ValidateDirectory validates all files in a directory
	ValidateDirectory(path string, format IaCFormat) error
	
	// ValidateSyntax validates syntax without file I/O
	ValidateSyntax(content string, format IaCFormat) error
	
	// GetValidationErrors returns detailed validation errors
	GetValidationErrors() []ValidationError
}

// ValidationError represents a validation error
type ValidationError struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Code     string `json:"code,omitempty"`
}

// DependencyAnalyzer defines the interface for analyzing resource dependencies
type DependencyAnalyzer interface {
	// AnalyzeDependencies analyzes dependencies between resources
	AnalyzeDependencies(resources []discovery.Resource) (map[string][]string, error)
	
	// GetDependencyGraph returns a dependency graph
	GetDependencyGraph(resources []discovery.Resource) (*DependencyGraph, error)
	
	// ValidateDependencies validates that dependencies are resolvable
	ValidateDependencies(dependencies map[string][]string) error
}

// DependencyGraph represents resource dependencies
type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

// DependencyNode represents a resource in the dependency graph
type DependencyNode struct {
	ID           string                 `json:"id"`
	ResourceType string                 `json:"resource_type"`
	Provider     discovery.CloudProvider `json:"provider"`
	Name         string                 `json:"name"`
	Level        int                    `json:"level"` // Dependency level (0 = no deps)
}

// DependencyEdge represents a dependency relationship
type DependencyEdge struct {
	From         string `json:"from"`           // Resource ID that depends
	To           string `json:"to"`             // Resource ID being depended on
	Type         string `json:"type"`           // Type of dependency
	Required     bool   `json:"required"`       // Whether dependency is required
	Attribute    string `json:"attribute,omitempty"` // Specific attribute dependency
}

// GenerationEngine defines the main interface for IaC generation
type GenerationEngine interface {
	// Generate generates IaC from discovered resources
	Generate(ctx context.Context, opts GenerationOptions) (*GenerationResult, error)
	
	// RegisterMapper registers a resource mapper for a cloud provider
	RegisterMapper(provider discovery.CloudProvider, mapper ResourceMapper)
	
	// RegisterGenerator registers a generator for an IaC format
	RegisterGenerator(format IaCFormat, generator Generator)
	
	// ValidateOptions validates generation options
	ValidateOptions(opts GenerationOptions) error
	
	// ListFormats returns supported IaC formats
	ListFormats() []IaCFormat
	
	// GetFormatCapabilities returns capabilities for a specific format
	GetFormatCapabilities(format IaCFormat) FormatCapabilities
	
	// Preview generates a preview of what would be generated
	Preview(ctx context.Context, opts GenerationOptions) (*PreviewResult, error)
}

// FormatCapabilities represents capabilities of an IaC format
type FormatCapabilities struct {
	Format            IaCFormat `json:"format"`
	SupportsModules   bool      `json:"supports_modules"`
	SupportsVariables bool      `json:"supports_variables"`
	SupportsOutputs   bool      `json:"supports_outputs"`
	SupportsState     bool      `json:"supports_state"`
	SupportsImports   bool      `json:"supports_imports"`
	SupportedProviders []discovery.CloudProvider `json:"supported_providers"`
}

// PreviewResult contains the results of a generation preview
type PreviewResult struct {
	ExpectedFiles    []FilePreview      `json:"expected_files"`
	ResourceCount    int                `json:"resource_count"`
	ProviderStats    map[discovery.CloudProvider]int `json:"provider_stats"`
	EstimatedLines   int                `json:"estimated_lines"`
	Warnings         []GenerationWarning `json:"warnings,omitempty"`
}

// FilePreview represents a preview of a file that would be generated
type FilePreview struct {
	Path         string   `json:"path"`
	Type         FileType `json:"type"`
	ResourceCount int     `json:"resource_count"`
	EstimatedSize int64   `json:"estimated_size"`
}
