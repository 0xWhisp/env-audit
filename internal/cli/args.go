package cli

import "fmt"

// Config holds parsed CLI arguments
type Config struct {
	FilePath   string   // --file path to .env file
	Required   []string // --required comma-separated required vars
	DumpMode   bool     // --dump output parsed config
	JSONOutput bool     // --json output results as JSON
	Quiet      bool     // --quiet/-q suppress stdout output
	Strict     bool     // --strict treat warnings as errors
	Help       bool     // --help show usage
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
		case "--quiet", "-q":
			cfg.Quiet = true
		case "--strict":
			cfg.Strict = true
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
