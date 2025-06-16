package steampipe

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/BigChiefRick/chimera/pkg/discovery"
)

const (
	defaultSteampipeHost = "localhost"
	defaultSteampipePort = 9193
	defaultDatabase      = "steampipe"
	defaultUser          = "steampipe"
)

// Connector implements the SteampipeConnector interface
type Connector struct {
	db     *sql.DB
	config Config
	logger *logrus.Logger
}

// Config contains Steampipe connection configuration
type Config struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
}

// NewConnector creates a new Steampipe connector
func NewConnector(config Config) *Connector {
	if config.Host == "" {
		config.Host = defaultSteampipeHost
	}
	if config.Port == 0 {
		config.Port = defaultSteampipePort
	}
	if config.Database == "" {
		config.Database = defaultDatabase
	}
	if config.User == "" {
		config.User = defaultUser
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Connector{
		config: config,
		logger: logrus.New(),
	}
}

// Connect establishes connection to Steampipe
func (c *Connector) Connect(ctx context.Context) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.config.Host, c.config.Port, c.config.User, c.config.Database, c.config.SSLMode)
	
	if c.config.Password != "" {
		connStr += fmt.Sprintf(" password=%s", c.config.Password)
	}

	var err error
	c.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to Steampipe: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping Steampipe database: %w", err)
	}

	c.logger.Info("Connected to Steampipe successfully")
	return nil
}

// Disconnect closes the connection to Steampipe
func (c *Connector) Disconnect() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Query executes a SQL query against Steampipe
func (c *Connector) Query(ctx context.Context, sql string) (*discovery.QueryResult, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to Steampipe")
	}

	rows, err := c.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare scan destinations
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var resultRows [][]interface{}
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to proper types
		row := make([]interface{}, len(columns))
		for i, v := range values {
			if v != nil {
				row[i] = v
			}
		}
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &discovery.QueryResult{
		Columns:  columns,
		Rows:     resultRows,
		RowCount: len(resultRows),
	}, nil
}

// ListTables returns available tables for the specified providers
func (c *Connector) ListTables(ctx context.Context, providers []discovery.CloudProvider) ([]discovery.TableInfo, error) {
	var tables []discovery.TableInfo
	
	for _, provider := range providers {
		providerTables, err := c.getProviderTables(ctx, provider)
		if err != nil {
			c.logger.Warnf("Failed to get tables for provider %s: %v", provider, err)
			continue
		}
		tables = append(tables, providerTables...)
	}

	return tables, nil
}

// getProviderTables gets tables for a specific provider
func (c *Connector) getProviderTables(ctx context.Context, provider discovery.CloudProvider) ([]discovery.TableInfo, error) {
	// Query information_schema to get tables for the provider
	query := `
		SELECT table_name, table_comment
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE $1
		ORDER BY table_name`

	pattern := string(provider) + "_%"
	rows, err := c.db.QueryContext(ctx, query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []discovery.TableInfo
	for rows.Next() {
		var tableName, tableComment sql.NullString
		if err := rows.Scan(&tableName, &tableComment); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		table := discovery.TableInfo{
			Name:        tableName.String,
			Provider:    provider,
			Description: tableComment.String,
		}

		// Get column information for this table
		columns, err := c.getTableColumns(ctx, tableName.String)
		if err != nil {
			c.logger.Warnf("Failed to get columns for table %s: %v", tableName.String, err)
		} else {
			table.Columns = columns
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// getTableColumns gets column information for a table
func (c *Connector) getTableColumns(ctx context.Context, tableName string) ([]discovery.ColumnInfo, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_comment
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = $1
		ORDER BY ordinal_position`

	rows, err := c.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []discovery.ColumnInfo
	for rows.Next() {
		var columnName, dataType, isNullable, columnComment sql.NullString
		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnComment); err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}

		column := discovery.ColumnInfo{
			Name:        columnName.String,
			Type:        dataType.String,
			Description: columnComment.String,
			Required:    isNullable.String == "NO",
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// GetSchema returns the schema for a specific table
func (c *Connector) GetSchema(ctx context.Context, table string) (*discovery.TableSchema, error) {
	// Determine provider from table name
	provider := c.getProviderFromTableName(table)
	
	// Get table info
	tables, err := c.getProviderTables(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider tables: %w", err)
	}

	for _, t := range tables {
		if t.Name == table {
			return &discovery.TableSchema{
				Table:   t,
				Columns: t.Columns,
			}, nil
		}
	}

	return nil, fmt.Errorf("table %s not found", table)
}

// getProviderFromTableName extracts provider from table name
func (c *Connector) getProviderFromTableName(tableName string) discovery.CloudProvider {
	parts := strings.Split(tableName, "_")
	if len(parts) > 0 {
		switch parts[0] {
		case "aws":
			return discovery.AWS
		case "azure":
			return discovery.Azure
		case "gcp":
			return discovery.GCP
		case "kubernetes":
			return discovery.Kubernetes
		}
	}
	return ""
}

// DiscoverResources discovers resources using Steampipe queries
func (c *Connector) DiscoverResources(ctx context.Context, providers []discovery.CloudProvider, resourceTypes []string) ([]discovery.Resource, error) {
	var allResources []discovery.Resource

	for _, provider := range providers {
		resources, err := c.discoverProviderResources(ctx, provider, resourceTypes)
		if err != nil {
			c.logger.Warnf("Failed to discover resources for provider %s: %v", provider, err)
			continue
		}
		allResources = append(allResources, resources...)
	}

	return allResources, nil
}

// discoverProviderResources discovers resources for a specific provider
func (c *Connector) discoverProviderResources(ctx context.Context, provider discovery.CloudProvider, resourceTypes []string) ([]discovery.Resource, error) {
	var resources []discovery.Resource

	// Get available tables for the provider
	tables, err := c.getProviderTables(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables for provider %s: %w", provider, err)
	}

	for _, table := range tables {
		// Skip if specific resource types are requested and this table doesn't match
		if len(resourceTypes) > 0 && !c.matchesResourceType(table.Name, resourceTypes) {
			continue
		}

		tableResources, err := c.queryTableResources(ctx, table, provider)
		if err != nil {
			c.logger.Warnf("Failed to query resources from table %s: %v", table.Name, err)
			continue
		}

		resources = append(resources, tableResources...)
	}

	return resources, nil
}

// matchesResourceType checks if a table name matches any of the requested resource types
func (c *Connector) matchesResourceType(tableName string, resourceTypes []string) bool {
	for _, resourceType := range resourceTypes {
		if strings.Contains(tableName, resourceType) {
			return true
		}
	}
	return false
}

// queryTableResources queries resources from a specific table
func (c *Connector) queryTableResources(ctx context.Context, table discovery.TableInfo, provider discovery.CloudProvider) ([]discovery.Resource, error) {
	// Build a basic query to get resources from this table
	query := c.buildResourceQuery(table)
	
	result, err := c.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table %s: %w", table.Name, err)
	}

	var resources []discovery.Resource
	for _, row := range result.Rows {
		resource := c.convertRowToResource(row, result.Columns, table, provider)
		resources = append(resources, resource)
	}

	return resources, nil
}

// buildResourceQuery builds a query to get resources from a table
func (c *Connector) buildResourceQuery(table discovery.TableInfo) string {
	// Start with basic fields that most tables have
	fields := []string{"name", "region"}
	
	// Add common fields if they exist in the table
	commonFields := []string{"id", "arn", "title", "tags", "zone", "project", "resource_group"}
	for _, field := range commonFields {
		if c.hasColumn(table, field) {
			fields = append(fields, field)
		}
	}

	return fmt.Sprintf("SELECT %s FROM %s LIMIT 1000", strings.Join(fields, ", "), table.Name)
}

// hasColumn checks if a table has a specific column
func (c *Connector) hasColumn(table discovery.TableInfo, columnName string) bool {
	for _, column := range table.Columns {
		if column.Name == columnName {
			return true
		}
	}
	return false
}

// convertRowToResource converts a query result row to a Resource
func (c *Connector) convertRowToResource(row []interface{}, columns []string, table discovery.TableInfo, provider discovery.CloudProvider) discovery.Resource {
	resource := discovery.Resource{
		Provider: provider,
		Type:     table.Name,
		Metadata: make(map[string]interface{}),
		Tags:     make(map[string]string),
	}

	// Map columns to resource fields
	for i, column := range columns {
		if i >= len(row) {
			break
		}

		value := row[i]
		if value == nil {
			continue
		}

		switch column {
		case "name", "title":
			if str, ok := value.(string); ok {
				resource.Name = str
			}
		case "id", "arn":
			if str, ok := value.(string); ok {
				resource.ID = str
			}
		case "region":
			if str, ok := value.(string); ok {
				resource.Region = str
			}
		case "zone":
			if str, ok := value.(string); ok {
				resource.Zone = str
			}
		case "project":
			if str, ok := value.(string); ok {
				resource.Project = str
			}
		case "resource_group":
			if str, ok := value.(string); ok {
				resource.ResourceGroup = str
			}
		case "tags":
			// Handle tags differently based on the data type
			if tags, ok := value.(map[string]interface{}); ok {
				for k, v := range tags {
					if str, ok := v.(string); ok {
						resource.Tags[k] = str
					}
				}
			}
		default:
			// Store other fields in metadata
			resource.Metadata[column] = value
		}
	}

	// Generate a unique ID if not provided
	if resource.ID == "" {
		resource.ID = fmt.Sprintf("%s-%s", resource.Type, resource.Name)
	}

	return resource
}