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

// OrganizationPattern defines how to organize generated code
type OrganizationPattern string

const (
	OrganizeByProvider     OrganizationPattern = "by_provider"
	OrganizeByService      OrganizationPattern = "by_service"
	OrganizeByRegion       OrganizationPattern = "by_region"
	OrganizeByResourceType OrganizationPattern = "by_resource_type"
	OrganizeFlat           OrganizationPattern = "flat"
)

// GenerationOptions contains configuration for IaC generation
type GenerationOptions struct {
	// Input
	Resources []discovery.Resource `json:"resources"`
	
	// Output format and organization
	Format       IaCFormat           `json:"format"`
	Organization OrganizationPattern `json:"organization"`
	
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
	Timestamp        string        `json:"timestamp,omitempty"`
}

// ModuleStructure defines how to organize generated code into modules
type ModuleStructure string

const (
	ModuleByProvider     ModuleStructure = "by_provider"
	ModuleByService      ModuleStructure = "by_service"
	ModuleByRegion       ModuleStructure = "by_region"
	ModuleByResourceType ModuleStructure = "by_resource_type"
	ModuleFlat           ModuleStructure = "flat"
	ModuleCustom         ModuleStructure = "custom"
)

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
	Checksum     string    `json:"checksum"`
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
	Organization   OrganizationPattern            `json:"organization"`
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
	Severity     ErrorSeverity          `json:"severity"`
	File         string                 `json:"file,omitempty"`
	Line         int                    `json:"line,omitempty"`
}

// GenerationWarning represents a warning during generation
type GenerationWarning struct {
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Provider     discovery.CloudProvider `json:"provider"`
	Message      string                 `json:"message"`
	Type         WarningType            `json:"type"`
	File         string                 `json:"file,omitempty"`
	Suggestion   string                 `json:"suggestion,omitempty"`
}

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// WarningType represents the type of warning
type WarningType string

const (
	WarningTypeDeprecated       WarningType = "deprecated"
	WarningTypeIncomplete       WarningType = "incomplete"
	WarningTypeUnsupported      WarningType = "unsupported"
	WarningTypeManualAction     WarningType = "manual_action"
	WarningTypeSecurityRisk     WarningType = "security_risk"
	WarningTypePerformanceRisk  WarningType = "performance_risk"
	WarningTypeBestPractice     WarningType = "best_practice"
	WarningTypeDataLoss         WarningType = "data_loss"
)

// TerraformResource represents a mapped Terraform resource
type TerraformResource struct {
	Type         string                    `json:"type"`           // e.g., "aws_vpc"
	Name         string                    `json:"name"`           // e.g., "main_vpc"
	Provider     discovery.CloudProvider   `json:"provider"`       // e.g., "aws"
	Config       map[string]interface{}    `json:"config"`         // Terraform configuration
	Dependencies []string                  `json:"dependencies"`   // Resource dependencies
	Outputs      map[string]string         `json:"outputs"`        // Generated outputs
	Variables    map[string]Variable       `json:"variables"`      // Required variables
	SourceInfo   SourceInfo                `json:"source_info"`    // Original discovery info
}

// MappedResource represents a discovered resource mapped to Terraform format
type MappedResource struct {
	OriginalResource discovery.Resource         `json:"original_resource"` // Original discovered resource
	ResourceType     string                    `json:"resource_type"`     // Terraform resource type (e.g., "aws_vpc")
	ResourceName     string                    `json:"resource_name"`     // Terraform resource name (e.g., "main_vpc")
	Configuration    map[string]interface{}    `json:"configuration"`     // Terraform configuration block
	Dependencies     []string                  `json:"dependencies"`      // List of resource dependencies
	Variables        map[string]Variable       `json:"variables"`         // Required variables for this resource
	Outputs          map[string]Output         `json:"outputs"`           // Outputs generated by this resource
}

// Variable represents a Terraform variable
type Variable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Sensitive   bool        `json:"sensitive"`
	Required    bool        `json:"required"`
}

// SourceInfo contains information about the original discovered resource
type SourceInfo struct {
	OriginalID       string                 `json:"original_id"`
	OriginalType     string                 `json:"original_type"`
	OriginalProvider discovery.CloudProvider `json:"original_provider"`
	OriginalRegion   string                 `json:"original_region"`
	DiscoveredAt     time.Time              `json:"discovered_at"`
	Metadata         map[string]interface{} `json:"metadata"`
	Tags             map[string]string      `json:"tags"`
}

// ProviderConfig represents provider configuration
type ProviderConfig struct {
	Name     string                 `json:"name"`     // e.g., "aws", "azurerm", "google"
	Source   string                 `json:"source"`   // e.g., "hashicorp/aws"
	Version  string                 `json:"version"`  // e.g., "~> 5.0"
	Config   map[string]interface{} `json:"config"`   // Provider-specific config
	Alias    string                 `json:"alias,omitempty"`
	Required bool                   `json:"required"`
}

// ModuleConfig represents a Terraform module configuration
type ModuleConfig struct {
	Name         string                 `json:"name"`
	Source       string                 `json:"source"`
	Version      string                 `json:"version,omitempty"`
	Providers    []ProviderConfig       `json:"providers"`
	Variables    map[string]Variable    `json:"variables"`
	Outputs      map[string]Output      `json:"outputs"`
	Resources    []TerraformResource    `json:"resources"`
	Dependencies []string               `json:"dependencies"`
}

// Output represents a Terraform output
type Output struct {
	Name        string      `json:"name"`
	Value       string      `json:"value"`       // Terraform expression
	Description string      `json:"description"`
	Sensitive   bool        `json:"sensitive"`
	Type        string      `json:"type,omitempty"`
}

// GenerationEngine is the main interface for IaC generation
type GenerationEngine interface {
	// Generate generates IaC from discovered resources
	Generate(ctx context.Context, opts GenerationOptions) (*GenerationResult, error)
	
	// ListFormats returns supported IaC formats
	ListFormats() []IaCFormat
	
	// ValidateOptions validates generation options
	ValidateOptions(opts GenerationOptions) error
	
	// GetFormatCapabilities returns capabilities for a specific format
	GetFormatCapabilities(format IaCFormat) FormatCapabilities
	
	// Preview generates a preview of what would be generated
	Preview(ctx context.Context, opts GenerationOptions) (*GenerationPreview, error)
}

// ResourceMapper defines the interface for mapping discovered resources to IaC resources
type ResourceMapper interface {
	// MapResource maps a discovered resource to an IaC resource
	MapResource(resource discovery.Resource) (*MappedResource, error)
	
	// GetProviderConfig returns the provider configuration needed
	GetProviderConfig(resources []discovery.Resource) (*ProviderConfig, error)
	
	// GetDependencies analyzes and returns resource dependencies
	GetDependencies(resource discovery.Resource, allResources []discovery.Resource) ([]string, error)
	
	// ValidateMapping validates that the mapping is correct
	ValidateMapping(original discovery.Resource, mapped MappedResource) error
	
	// GetSupportedTypes returns the resource types this mapper supports
	GetSupportedTypes() []string
	
	// Provider returns the cloud provider this mapper supports
	Provider() discovery.CloudProvider
}

// TerraformGenerator defines the interface for Terraform-specific generation
type TerraformGenerator interface {
	// GenerateResource generates Terraform HCL for a single resource
	GenerateResource(resource TerraformResource) (string, error)
	
	// GenerateResourceHCL generates HCL for a mapped resource
	GenerateResourceHCL(resource MappedResource) (string, error)
	
	// GenerateProvider generates provider configuration block
	GenerateProvider(config ProviderConfig) (string, error)
	
	// GenerateVariables generates variables.tf content
	GenerateVariables(variables map[string]Variable) (string, error)
	
	// GenerateOutputs generates outputs.tf content
	GenerateOutputs(outputs map[string]Output) (string, error)
	
	// GenerateVersions generates versions.tf content
	GenerateVersions(providers []ProviderConfig) (string, error)
	
	// GenerateModule generates a complete module
	GenerateModule(config ModuleConfig) (map[string]string, error)
	
	// ValidateSyntax validates generated Terraform syntax
	ValidateSyntax(content string) error
}

// TemplateEngine defines the interface for template-based generation
type TemplateEngine interface {
	// Render renders a template with the provided data
	Render(templateName string, data interface{}) (string, error)
	
	// RegisterTemplate registers a new template
	RegisterTemplate(name string, template string) error
	
	// ListTemplates returns available templates
	ListTemplates() []string
	
	// ValidateTemplate validates a template
	ValidateTemplate(template string) error
	
	// GetTemplate returns a template by name
	GetTemplate(name string) (string, error)
}

// FormatCapabilities describes what a specific IaC format supports
type FormatCapabilities struct {
	Format               IaCFormat                   `json:"format"`
	SupportedProviders   []discovery.CloudProvider   `json:"supported_providers"`
	SupportedResources   map[string][]string         `json:"supported_resources"`
	SupportsModules      bool                        `json:"supports_modules"`
	SupportsState        bool                        `json:"supports_state"`
	SupportsVariables    bool                        `json:"supports_variables"`
	SupportsOutputs      bool                        `json:"supports_outputs"`
	SupportsValidation   bool                        `json:"supports_validation"`
	SupportsImports      bool                        `json:"supports_imports"`
	OrganizationPatterns []OrganizationPattern       `json:"organization_patterns"`
}

// GenerationPreview provides a preview of generation results
type GenerationPreview struct {
	FileStructure    []PreviewFile         `json:"file_structure"`
	ResourceCount    int                   `json:"resource_count"`
	EstimatedSize    int64                 `json:"estimated_size"`
	UnsupportedItems []UnsupportedResource `json:"unsupported_items"`
	Warnings         []GenerationWarning   `json:"warnings"`
	Providers        []ProviderConfig      `json:"providers"`
	Variables        map[string]Variable   `json:"variables"`
	Outputs          map[string]Output     `json:"outputs"`
}

// PreviewFile represents a file in the generation preview
type PreviewFile struct {
	Path          string   `json:"path"`
	Type          FileType `json:"type"`
	ResourceCount int      `json:"resource_count"`
	EstimatedSize int64    `json:"estimated_size"`
	Dependencies  []string `json:"dependencies,omitempty"`
}

// UnsupportedResource represents a resource that cannot be generated
type UnsupportedResource struct {
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Provider     discovery.CloudProvider `json:"provider"`
	Reason       string                 `json:"reason"`
	Suggestion   string                 `json:"suggestion,omitempty"`
}

// FileOrganizer defines the interface for organizing generated files
type FileOrganizer interface {
	// OrganizeFiles organizes resources into file structure based on pattern
	OrganizeFiles(resources []TerraformResource, pattern OrganizationPattern) (map[string][]TerraformResource, error)
	
	// GetFilePath returns the file path for a resource
	GetFilePath(resource TerraformResource, pattern OrganizationPattern) (string, error)
	
	// ValidateOrganization validates the organization pattern
	ValidateOrganization(pattern OrganizationPattern, resources []TerraformResource) error
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
