package cli

import "fmt"

// Config holds parsed CLI arguments
type Config struct {
	FilePath     string   // --file path to .env file
	Required     []string // --required comma-separated required vars
	ExampleFile  string   // --example path to .env.example file
	DiffFile     string   // --diff path to second file for comparison
	Ignore       []string // --ignore comma-separated keys to ignore
	DumpMode     bool     // --dump output parsed config
	JSONOutput   bool     // --json output results as JSON
	GitHubOutput bool     // --github output results in GitHub Actions format
	Quiet        bool     // --quiet/-q suppress stdout output
	Strict       bool     // --strict treat warnings as errors
	CheckLeaks   bool     // --check-leaks analyze values for secret patterns
	NoColor      bool     // --no-color disable colored output
	Watch        bool     // --watch watch file for changes
	Init         bool     // --init generate .env.example file
	Force        bool     // --force overwrite existing files
	Help         bool     // --help show usage
	Version      bool     // --version/-v show version
}

// ParseArgs parses command line arguments into Config
func ParseArgs(args []string) (*Config, error) {
	cfg := &Config{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			cfg.Help = true
		case "--dump", "-d":
			cfg.DumpMode = true
		case "--json":
			cfg.JSONOutput = true
		case "--github":
			cfg.GitHubOutput = true
		case "--quiet", "-q":
			cfg.Quiet = true
		case "--strict":
			cfg.Strict = true
		case "--check-leaks":
			cfg.CheckLeaks = true
		case "--init":
			cfg.Init = true
		case "--force":
			cfg.Force = true
		case "--no-color":
			cfg.NoColor = true
		case "--watch", "-w":
			cfg.Watch = true
		case "--version", "-V":
			cfg.Version = true
		case "--file", "-f":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			i++
			cfg.FilePath = args[i]
		case "--required", "-r":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			i++
			cfg.Required = parseCommaSeparated(args[i])
		case "--example", "-e":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			i++
			cfg.ExampleFile = args[i]
		case "--diff":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			i++
			cfg.DiffFile = args[i]
		case "--ignore", "-i":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			i++
			cfg.Ignore = parseCommaSeparated(args[i])
		default:
			return nil, fmt.Errorf("unknown argument: %s", arg)
		}
	}

	return cfg, nil
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, part := range splitComma(s) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitComma(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// MergeWithFileConfig merges file config into CLI config
// CLI flags take precedence over file config
func (cfg *Config) MergeWithFileConfig(file *FileConfig) {
	if file == nil {
		return
	}

	// Only apply file config values if CLI didn't set them
	if cfg.FilePath == "" && file.File != "" {
		cfg.FilePath = file.File
	}
	if len(cfg.Required) == 0 && len(file.Required) > 0 {
		cfg.Required = file.Required
	}
	if cfg.ExampleFile == "" && file.Example != "" {
		cfg.ExampleFile = file.Example
	}
	if len(cfg.Ignore) == 0 && len(file.Ignore) > 0 {
		cfg.Ignore = file.Ignore
	}

	// Boolean flags: file config only sets if CLI didn't enable
	if !cfg.Strict && file.Strict {
		cfg.Strict = true
	}
	if !cfg.CheckLeaks && file.CheckLeaks {
		cfg.CheckLeaks = true
	}
	if !cfg.Quiet && file.Quiet {
		cfg.Quiet = true
	}
	if !cfg.JSONOutput && file.JSON {
		cfg.JSONOutput = true
	}
	if !cfg.GitHubOutput && file.GitHub {
		cfg.GitHubOutput = true
	}
	if !cfg.NoColor && file.NoColor {
		cfg.NoColor = true
	}
}

// FileConfig holds config loaded from file
type FileConfig struct {
	File       string
	Required   []string
	Example    string
	Ignore     []string
	Strict     bool
	CheckLeaks bool
	Quiet      bool
	JSON       bool
	GitHub     bool
	NoColor    bool
}
