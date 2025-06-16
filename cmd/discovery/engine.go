package discovery

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Engine implements the DiscoveryEngine interface
type Engine struct {
	connectors map[CloudProvider]ProviderConnector
	steampipe  SteampipeConnector
	logger     *logrus.Logger
	config     EngineConfig
}

// EngineConfig contains configuration for the discovery engine
type EngineConfig struct {
	MaxConcurrency int           `yaml:"max_concurrency" json:"max_concurrency"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	RetryAttempts  int           `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay     time.Duration `yaml:"retry_delay" json:"retry_delay"`
}

// NewEngine creates a new discovery engine
func NewEngine(config EngineConfig, steampipeConnector SteampipeConnector) *Engine {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 10
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Minute
	}
	if config.RetryAttempts <= 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &Engine{
		connectors: make(map[CloudProvider]ProviderConnector),
		steampipe:  steampipeConnector,
		logger:     logrus.New(),
		config:     config,
	}
}

// RegisterConnector registers a provider connector
func (e *Engine) RegisterConnector(connector ProviderConnector) {
	e.connectors[connector.Provider()] = connector
	e.logger.Infof("Registered connector for provider: %s", connector.Provider())
}

// Discover discovers resources based on the provided options
func (e *Engine) Discover(ctx context.Context, opts DiscoveryOptions) (*DiscoveryResult, error) {
	startTime := time.Now()
	
	// Validate options
	if err := e.validateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid discovery options: %w", err)
	}

	// Apply timeout to context
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Initialize result
	result := &DiscoveryResult{
		Resources: make([]Resource, 0),
		Errors:    make([]DiscoveryError, 0),
		Metadata: DiscoveryMetadata{
			StartTime:     startTime,
			ProviderStats: make(map[string]int),
			Filters:       opts.Filters,
		},
	}

	// Determine which discovery method to use
	if e.steampipe != nil && e.shouldUseSteampipe(opts) {
		return e.discoverWithSteampipe(ctx, opts, result)
	}

	return e.discoverWithConnectors(ctx, opts, result)
}

// discoverWithSteampipe uses Steampipe for unified discovery
func (e *Engine) discoverWithSteampipe(ctx context.Context, opts DiscoveryOptions, result *DiscoveryResult) (*DiscoveryResult, error) {
	e.logger.Info("Using Steampipe for resource discovery")

	// Connect to Steampipe
	if err := e.steampipe.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Steampipe: %w", err)
	}
	defer func() {
		if err := e.steampipe.Disconnect(); err != nil {
			e.logger.Warnf("Failed to disconnect from Steampipe: %v", err)
		}
	}()

	// Discover resources using Steampipe
	resources, err := e.steampipe.DiscoverResources(ctx, opts.Providers, opts.ResourceTypes)
	if err != nil {
		result.Errors = append(result.Errors, DiscoveryError{
			Message: fmt.Sprintf("Steampipe discovery failed: %v", err),
			Error:   err,
		})
	} else {
		result.Resources = append(result.Resources, resources...)
	}

	// Apply filters
	result.Resources = e.applyFilters(result.Resources, opts.Filters)

	// Finalize metadata
	e.finalizeMetadata(result)

	return result, nil
}

// discoverWithConnectors uses individual provider connectors
func (e *Engine) discoverWithConnectors(ctx context.Context, opts DiscoveryOptions, result *DiscoveryResult) (*DiscoveryResult, error) {
	e.logger.Info("Using provider connectors for resource discovery")

	// Create a channel for results
	resourceChan := make(chan []Resource, len(opts.Providers))
	errorChan := make(chan DiscoveryError, len(opts.Providers))

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, opts.MaxConcurrency)
	if opts.MaxConcurrency <= 0 {
		opts.MaxConcurrency = e.config.MaxConcurrency
	}

	var wg sync.WaitGroup

	// Discover resources from each provider concurrently
	for _, provider := range opts.Providers {
		wg.Add(1)
		go func(provider CloudProvider) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			e.discoverFromProvider(ctx, provider, opts, resourceChan, errorChan)
		}(provider)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(resourceChan)
		close(errorChan)
	}()

	// Collect results
	for resources := range resourceChan {
		result.Resources = append(result.Resources, resources...)
	}

	// Collect errors
	for err := range errorChan {
		result.Errors = append(result.Errors, err)
	}

	// Apply filters
	result.Resources = e.applyFilters(result.Resources, opts.Filters)

	// Finalize metadata
	e.finalizeMetadata(result)

	return result, nil
}

// discoverFromProvider discovers resources from a single provider
func (e *Engine) discoverFromProvider(ctx context.Context, provider CloudProvider, opts DiscoveryOptions, resourceChan chan<- []Resource, errorChan chan<- DiscoveryError) {
	connector, exists := e.connectors[provider]
	if !exists {
		errorChan <- DiscoveryError{
			Provider: provider,
			Message:  fmt.Sprintf("No connector available for provider: %s", provider),
		}
		return
	}

	// Validate credentials
	if err := connector.ValidateCredentials(ctx); err != nil {
		errorChan <- DiscoveryError{
			Provider: provider,
			Message:  fmt.Sprintf("Credential validation failed: %v", err),
			Error:    err,
		}
		return
	}

	// Prepare provider-specific options
	providerOpts := ProviderDiscoveryOptions{
		Regions:       opts.Regions,
		ResourceTypes: opts.ResourceTypes,
		Filters:       opts.Filters,
	}

	// Discover resources with retry logic
	var resources []Resource
	var err error

	for attempt := 0; attempt < e.config.RetryAttempts; attempt++ {
		resources, err = connector.Discover(ctx, providerOpts)
		if err == nil {
			break
		}

		if attempt < e.config.RetryAttempts-1 {
			e.logger.Warnf("Discovery attempt %d failed for provider %s, retrying: %v", attempt+1, provider, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(e.config.RetryDelay):
			}
		}
	}

	if err != nil {
		errorChan <- DiscoveryError{
			Provider: provider,
			Message:  fmt.Sprintf("Discovery failed after %d attempts: %v", e.config.RetryAttempts, err),
			Error:    err,
		}
		return
	}

	resourceChan <- resources
}

// validateOptions validates discovery options
func (e *Engine) validateOptions(opts DiscoveryOptions) error {
	if len(opts.Providers) == 0 {
		return fmt.Errorf("at least one provider must be specified")
	}

	// Check if all requested providers are supported
	for _, provider := range opts.Providers {
		if _, exists := e.connectors[provider]; !exists && e.steampipe == nil {
			return fmt.Errorf("unsupported provider: %s", provider)
		}
	}

	return nil
}

// shouldUseSteampipe determines if Steampipe should be used for discovery
func (e *Engine) shouldUseSteampipe(opts DiscoveryOptions) bool {
	// Use Steampipe if available and multiple providers are requested
	return len(opts.Providers) > 1
}

// applyFilters applies filters to discovered resources
func (e *Engine) applyFilters(resources []Resource, filters []Filter) []Resource {
	if len(filters) == 0 {
		return resources
	}

	var filtered []Resource
	for _, resource := range resources {
		if e.matchesFilters(resource, filters) {
			filtered = append(filtered, resource)
		}
	}

	return filtered
}

// matchesFilters checks if a resource matches all filters
func (e *Engine) matchesFilters(resource Resource, filters []Filter) bool {
	for _, filter := range filters {
		if !e.matchesFilter(resource, filter) {
			return false
		}
	}
	return true
}

// matchesFilter checks if a resource matches a single filter
func (e *Engine) matchesFilter(resource Resource, filter Filter) bool {
	var value interface{}

	// Get the field value from the resource
	switch filter.Field {
	case "name":
		value = resource.Name
	case "type":
		value = resource.Type
	case "provider":
		value = string(resource.Provider)
	case "region":
		value = resource.Region
	case "zone":
		value = resource.Zone
	default:
		// Check in metadata
		if metaValue, exists := resource.Metadata[filter.Field]; exists {
			value = metaValue
		} else if tagValue, exists := resource.Tags[filter.Field]; exists {
			value = tagValue
		} else if strings.HasPrefix(filter.Field, "tags.") {
			tagKey := strings.TrimPrefix(filter.Field, "tags.")
			if tagValue, exists := resource.Tags[tagKey]; exists {
				value = tagValue
			} else {
				// Tag doesn't exist
				if filter.Operator == "exists" {
					return filter.Value == false
				}
				return false
			}
		} else {
			return false
		}
	}

	// Apply the filter operator
	switch filter.Operator {
	case "eq", "=", "==":
		return value == filter.Value
	case "ne", "!=":
		return value != filter.Value
	case "contains":
		if str, ok := value.(string); ok {
			if filterStr, ok := filter.Value.(string); ok {
				return strings.Contains(str, filterStr)
			}
		}
		return false
	case "in":
		if filterSlice, ok := filter.Value.([]interface{}); ok {
			for _, filterValue := range filterSlice {
				if value == filterValue {
					return true
				}
			}
		}
		return false
	case "exists":
		if expectedExists, ok := filter.Value.(bool); ok {
			return expectedExists == (value != nil)
		}
		return value != nil
	default:
		e.logger.Warnf("Unknown filter operator: %s", filter.Operator)
		return true
	}
}

// finalizeMetadata calculates final metadata for the discovery result
func (e *Engine) finalizeMetadata(result *DiscoveryResult) {
	result.Metadata.EndTime = time.Now()
	result.Metadata.Duration = result.Metadata.EndTime.Sub(result.Metadata.StartTime)
	result.Metadata.ResourceCount = len(result.Resources)
	result.Metadata.ErrorCount = len(result.Errors)

	// Calculate provider statistics
	for _, resource := range result.Resources {
		providerKey := string(resource.Provider)
		result.Metadata.ProviderStats[providerKey]++
	}
}

// ListProviders returns the list of supported providers
func (e *Engine) ListProviders() []CloudProvider {
	providers := make([]CloudProvider, 0, len(e.connectors))
	for provider := range e.connectors {
		providers = append(providers, provider)
	}
	return providers
}

// ValidateCredentials validates credentials for the specified providers
func (e *Engine) ValidateCredentials(ctx context.Context, providers []CloudProvider) error {
	for _, provider := range providers {
		connector, exists := e.connectors[provider]
		if !exists {
			return fmt.Errorf("no connector available for provider: %s", provider)
		}

		if err := connector.ValidateCredentials(ctx); err != nil {
			return fmt.Errorf("credential validation failed for provider %s: %w", provider, err)
		}
	}

	return nil
}

// GetProviderRegions returns available regions for a provider
func (e *Engine) GetProviderRegions(ctx context.Context, provider CloudProvider) ([]string, error) {
	connector, exists := e.connectors[provider]
	if !exists {
		return nil, fmt.Errorf("no connector available for provider: %s", provider)
	}

	return connector.GetRegions(ctx)
}

// GetResourceTypes returns available resource types for a provider
func (e *Engine) GetResourceTypes(ctx context.Context, provider CloudProvider) ([]string, error) {
	connector, exists := e.connectors[provider]
	if !exists {
		return nil, fmt.Errorf("no connector available for provider: %s", provider)
	}

	return connector.GetResourceTypes(ctx)
}
