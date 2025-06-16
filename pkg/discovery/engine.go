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

	// For Phase 1, return framework demonstration
	result.Metadata.EndTime = time.Now()
	result.Metadata.Duration = result.Metadata.EndTime.Sub(result.Metadata.StartTime)
	result.Metadata.ResourceCount = 0

	return result, nil
}

// validateOptions validates discovery options
func (e *Engine) validateOptions(opts DiscoveryOptions) error {
	if len(opts.Providers) == 0 {
		return fmt.Errorf("at least one provider must be specified")
	}
	return nil
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
