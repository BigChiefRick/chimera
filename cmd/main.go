package main

import (
	"fmt"
	"os"
	"strings"
)

var version = "v0.1.0-dev"

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "--help", "-h", "help":
			showHelp()
			return
		case "--version", "-v", "version":
			showVersion()
			return
		case "discover":
			handleDiscover()
			return
		case "generate":
			handleGenerate()
			return
		case "config":
			handleConfig()
			return
		}
	}

	// Default behavior
	fmt.Printf("Chimera %s - Multi-cloud infrastructure discovery and IaC generation tool\n", version)
	fmt.Println("Use --help for usage information")
}

func showHelp() {
	fmt.Printf(`Chimera %s - Multi-cloud infrastructure discovery and IaC generation tool

USAGE:
    chimera [COMMAND] [OPTIONS]

COMMANDS:
    discover    Discover infrastructure resources from cloud providers
    generate    Generate Infrastructure as Code from discovered resources  
    config      Manage configuration settings
    version     Show version information
    help        Show this help message

GLOBAL OPTIONS:
    --verbose, -v     Enable verbose output
    --debug           Enable debug output
    --config PATH     Configuration file path
    --help, -h        Show help
    --version         Show version

EXAMPLES:
    # Discover AWS resources in us-east-1
    chimera discover --provider aws --region us-east-1

    # Generate Terraform from discovered resources
    chimera generate --format terraform --output ./infrastructure

    # Show configuration
    chimera config show

For more information about a specific command, use:
    chimera [COMMAND] --help

Documentation: https://github.com/BigChiefRick/chimera
`, version)
}

func showVersion() {
	fmt.Printf("Chimera %s\n", version)
	fmt.Println("Built with Go")
	fmt.Println("https://github.com/BigChiefRick/chimera")
}

func handleDiscover() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		fmt.Printf(`chimera discover - Discover infrastructure resources

USAGE:
    chimera discover [OPTIONS]

OPTIONS:
    --provider PROVIDERS    Cloud providers (aws,azure,gcp,vmware,kvm)
    --region REGIONS       Regions to scan
    --resource-type TYPES  Resource types to discover
    --output PATH          Output file path
    --format FORMAT        Output format (json,yaml,table)
    --steampipe           Use Steampipe for discovery
    --filter FILTERS      Apply filters
    --help, -h            Show this help

EXAMPLES:
    chimera discover --provider aws --region us-east-1
    chimera discover --provider aws,azure --output resources.json
    chimera discover --steampipe --provider aws,gcp --format yaml
`)
		return
	}

	fmt.Println("🔍 Discovery command (implementation in progress)")
	fmt.Println("This command will discover infrastructure resources from cloud providers")
	fmt.Println("Use --help for detailed usage information")
}

func handleGenerate() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		fmt.Printf(`chimera generate - Generate Infrastructure as Code

USAGE:
    chimera generate [OPTIONS]

OPTIONS:
    --input PATH          Input file with discovered resources
    --format FORMAT       IaC format (terraform,pulumi,cloudformation)
    --output PATH         Output directory
    --organize-by TYPE    Organization method (provider,service,region)
    --help, -h           Show this help

EXAMPLES:
    chimera generate --input resources.json --format terraform
    chimera generate --format pulumi --output ./infrastructure
`)
		return
	}

	fmt.Println("⚙️  Generation command (implementation in progress)")
	fmt.Println("This command will generate Infrastructure as Code from discovered resources")
	fmt.Println("Use --help for detailed usage information")
}

func handleConfig() {
	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "--help", "-h":
			fmt.Printf(`chimera config - Manage configuration

USAGE:
    chimera config [SUBCOMMAND]

SUBCOMMANDS:
    init        Initialize default configuration
    show        Show current configuration
    validate    Validate configuration
    help        Show this help

EXAMPLES:
    chimera config init
    chimera config show
    chimera config validate
`)
			return
		case "show":
			fmt.Println("📋 Current configuration (default values):")
			fmt.Println("Config file: ~/.chimera.yaml (not found, using defaults)")
			fmt.Println("Debug: false")
			fmt.Println("Verbose: false")
			fmt.Println("Output format: json")
			fmt.Println("Use 'chimera config init' to create a configuration file")
			return
		case "init":
			fmt.Println("📝 Configuration initialization (implementation in progress)")
			fmt.Println("This will create a default configuration file at ~/.chimera.yaml")
			return
		case "validate":
			fmt.Println("✅ Configuration validation (using defaults)")
			fmt.Println("No configuration file found, defaults are valid")
			return
		}
	}

	fmt.Println("⚙️  Configuration management")
	fmt.Println("Use 'chimera config --help' for available commands")
}