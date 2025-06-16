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
	Timeout          time.Duration `json:"timeout"`
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
	FileTypeMain      FileType = "main"
	FileTypeVariables FileType = "variables"
	FileTypeOutputs   FileType = "outputs"
	FileTypeProvider  FileType = "provider"
	FileTypeState     FileType = "state"
	FileTypeModule    FileType = "module"
	FileTypeData      FileType = "data"
)

// GenerationMetadata contains metadata about the generation operation
type GenerationMetadata struct {
	StartTime      time.Time                    `json:"start_time"`
	EndTime        time.Time                    `json:"end_time"`
	Duration       time.Duration                `json:"duration"`
	ResourceCount  int                          `json:"resource_count"`
	FileCount      int                          `json:"file_count"`
	LinesGenerated int                          `json:"lines_generated"`
	Format         IaCFormat                    `json:"format"`
	ProviderStats  map[discovery.CloudProvider]int `json:"provider_stats"`
	ErrorCount     int                          `json:"error_count"`
	WarningCount   int                          `json:"warning_count"`
}

// GenerationError represents an error during generation
type GenerationError struct {
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Provider     discovery.CloudProvider `json:"provider"`
	Message      string                 `json:"message"`
	Error        error                  `json:"-"`
	Severity     ErrorSeverity          `json:"severity"`
}

// GenerationWarning represents a warning during generation
type GenerationWarning struct {
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Provider     discovery.CloudProvider `json:"provider"`
	Message      string                 `json:"message"`
	Type         WarningType            `json:"type"`
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
)

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

// FormatCapabilities describes what a specific IaC format supports
type FormatCapabilities struct {
	Format              IaCFormat                   `json:"format"`
	SupportedProviders  []discovery.CloudProvider   `json:"supported_providers"`
	SupportedResources  map[string][]string         `json:"supported_resources"`
	SupportsModules     bool                        `json:"supports_modules"`
	SupportsState       bool                        `json:"supports_state"`
	SupportsVariables   bool                        `json:"supports_variables"`
	SupportsOutputs     bool                        `json:"supports_outputs"`
	SupportsValidation  bool                        `json:"supports_validation"`
}

// GenerationPreview provides a preview of generation results
type GenerationPreview struct {
	FileStructure    []PreviewFile         `json:"file_structure"`
	ResourceCount    int                   `json:"resource_count"`
	EstimatedSize    int64                 `json:"estimated_size"`
	UnsupportedItems []UnsupportedResource `json:"unsupported_items"`
	Warnings         []GenerationWarning   `json:"warnings"`
}

// PreviewFile represents a file in the generation preview
type PreviewFile struct {
	Path          string   `json:"path"`
	Type          FileType `json:"type"`
	ResourceCount int      `json:"resource_count"`
	EstimatedSize int64    `json:"estimated_size"`
}

// UnsupportedResource represents a resource that cannot be generated
type UnsupportedResource struct {
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Provider     discovery.CloudProvider `json:"provider"`
	Reason       string                 `json:"reason"`
}

// TerraformerConnector defines the interface for Terraformer integration
type TerraformerConnector interface {
	// Generate generates Terraform code using Terraformer
	Generate(ctx context.Context, opts TerraformerOptions) (*GenerationResult, error)
	
	// ListProviders returns supported providers
	ListProviders() []discovery.CloudProvider
	
	// ListResources returns supported resources for a provider
	ListResources(provider discovery.CloudProvider) ([]string, error)
	
	// ValidateInstallation validates that Terraformer is properly installed
	ValidateInstallation() error
}

// TerraformerOptions contains Terraformer-specific options
type TerraformerOptions struct {
	Provider        discovery.CloudProvider `json:"provider"`
	Resources       []string                `json:"resources"`
	Regions         []string                `json:"regions,omitempty"`
	Filters         []string                `json:"filters,omitempty"`
	OutputPath      string                  `json:"output_path"`
	Connect         bool                    `json:"connect"`
	Compact         bool                    `json:"compact"`
	Verbose         bool                    `json:"verbose"`
	Excludes        []string                `json:"excludes,omitempty"`
	PathPattern     string                  `json:"path_pattern,omitempty"`
	State           string                  `json:"state"` // local or bucket
	Bucket          string                  `json:"bucket,omitempty"`
	RetryNumber     int                     `json:"retry_number"`
	RetrySleepMs    int                     `json:"retry_sleep_ms"`
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
}