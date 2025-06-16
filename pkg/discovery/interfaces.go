package discovery

import (
	"context"
	"time"
)

// CloudProvider represents the different cloud providers supported
type CloudProvider string

const (
	AWS        CloudProvider = "aws"
	Azure      CloudProvider = "azure"
	GCP        CloudProvider = "gcp"
	VMware     CloudProvider = "vmware"
	KVM        CloudProvider = "kvm"
	Kubernetes CloudProvider = "kubernetes"
)

// Resource represents a discovered cloud resource
type Resource struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Provider     CloudProvider          `json:"provider"`
	Region       string                 `json:"region,omitempty"`
	Zone         string                 `json:"zone,omitempty"`
	ResourceGroup string                `json:"resource_group,omitempty"`
	Project      string                 `json:"project,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	Tags         map[string]string      `json:"tags,omitempty"`
	CreatedAt    *time.Time             `json:"created_at,omitempty"`
	UpdatedAt    *time.Time             `json:"updated_at,omitempty"`
}

// DiscoveryOptions contains configuration for resource discovery
type DiscoveryOptions struct {
	Providers      []CloudProvider `json:"providers"`
	Regions        []string        `json:"regions,omitempty"`
	ResourceTypes  []string        `json:"resource_types,omitempty"`
	IncludeTags    []string        `json:"include_tags,omitempty"`
	ExcludeTags    []string        `json:"exclude_tags,omitempty"`
	MaxConcurrency int             `json:"max_concurrency"`
	Timeout        time.Duration   `json:"timeout"`
	Filters        []Filter        `json:"filters,omitempty"`
}

// Filter represents a resource filter
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, in, contains, etc.
	Value    interface{} `json:"value"`
}

// DiscoveryResult contains the results of a discovery operation
type DiscoveryResult struct {
	Resources []Resource        `json:"resources"`
	Errors    []DiscoveryError  `json:"errors,omitempty"`
	Metadata  DiscoveryMetadata `json:"metadata"`
}

// DiscoveryError represents an error during discovery
type DiscoveryError struct {
	Provider    CloudProvider `json:"provider"`
	Region      string        `json:"region,omitempty"`
	ResourceType string       `json:"resource_type,omitempty"`
	Message     string        `json:"message"`
	Error       error         `json:"-"`
}

// DiscoveryMetadata contains metadata about the discovery operation
type DiscoveryMetadata struct {
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	Duration      time.Duration     `json:"duration"`
	ResourceCount int               `json:"resource_count"`
	ProviderStats map[string]int    `json:"provider_stats"`
	ErrorCount    int               `json:"error_count"`
	Filters       []Filter          `json:"filters_applied"`
}

// DiscoveryEngine is the main interface for resource discovery
type DiscoveryEngine interface {
	// Discover discovers resources based on the provided options
	Discover(ctx context.Context, opts DiscoveryOptions) (*DiscoveryResult, error)
	
	// ListProviders returns the list of supported providers
	ListProviders() []CloudProvider
	
	// ValidateCredentials validates credentials for the specified providers
	ValidateCredentials(ctx context.Context, providers []CloudProvider) error
	
	// GetProviderRegions returns available regions for a provider
	GetProviderRegions(ctx context.Context, provider CloudProvider) ([]string, error)
	
	// GetResourceTypes returns available resource types for a provider
	GetResourceTypes(ctx context.Context, provider CloudProvider) ([]string, error)
}

// ProviderConnector represents a connector for a specific cloud provider
type ProviderConnector interface {
	// Provider returns the cloud provider this connector supports
	Provider() CloudProvider
	
	// Discover discovers resources for this provider
	Discover(ctx context.Context, opts ProviderDiscoveryOptions) ([]Resource, error)
	
	// ValidateCredentials validates credentials for this provider
	ValidateCredentials(ctx context.Context) error
	
	// GetRegions returns available regions for this provider
	GetRegions(ctx context.Context) ([]string, error)
	
	// GetResourceTypes returns available resource types for this provider
	GetResourceTypes(ctx context.Context) ([]string, error)
}

// ProviderDiscoveryOptions contains provider-specific discovery options
type ProviderDiscoveryOptions struct {
	Regions       []string          `json:"regions,omitempty"`
	ResourceTypes []string          `json:"resource_types,omitempty"`
	Filters       []Filter          `json:"filters,omitempty"`
	Credentials   map[string]string `json:"credentials,omitempty"`
	ExtraParams   map[string]interface{} `json:"extra_params,omitempty"`
}

// SteampipeConnector defines the interface for Steampipe integration
type SteampipeConnector interface {
	// Connect establishes connection to Steampipe
	Connect(ctx context.Context) error
	
	// Disconnect closes the connection to Steampipe
	Disconnect() error
	
	// Query executes a SQL query against Steampipe
	Query(ctx context.Context, sql string) (*QueryResult, error)
	
	// ListTables returns available tables for the specified providers
	ListTables(ctx context.Context, providers []CloudProvider) ([]TableInfo, error)
	
	// GetSchema returns the schema for a specific table
	GetSchema(ctx context.Context, table string) (*TableSchema, error)
}

// QueryResult represents the result of a Steampipe query
type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	RowCount int            `json:"row_count"`
}

// TableInfo contains information about a Steampipe table
type TableInfo struct {
	Name        string        `json:"name"`
	Provider    CloudProvider `json:"provider"`
	Description string        `json:"description"`
	Columns     []ColumnInfo  `json:"columns"`
}

// ColumnInfo contains information about a table column
type ColumnInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// TableSchema represents the schema of a Steampipe table
type TableSchema struct {
	Table   TableInfo    `json:"table"`
	Columns []ColumnInfo `json:"columns"`
}