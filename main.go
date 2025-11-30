package main

import (
	"fmt"
	"io"
	"os"
)

// Config holds parsed CLI arguments
type Config struct {
	FilePath string   // --file path to .env file
	Required []string // --required comma-separated required vars
	DumpMode bool     // --dump output parsed config
	Help     bool     // --help show usage
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the main logic and returns the exit code
func run(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}

	if cfg.Help {
		printUsage(stdout)
		return 0
	}

	var env map[string]string
	var duplicates []string

	if cfg.FilePath != "" {
		result, err := ParseEnvFile(cfg.FilePath)
		if err != nil {
			fmt.Fprintln(stderr, "Error:", err)
			return 2
		}
		env = result.Entries
		duplicates = result.Duplicates
	} else {
		env = ReadEnv()
	}

	if cfg.DumpMode {
		fmt.Fprintln(stdout, FormatConfig(env))
		return 0
	}

	scanResult := Scan(env, cfg.Required, duplicates)
	fmt.Fprint(stdout, FormatSummary(scanResult))

	if scanResult.HasRisks {
		return 1
	}
	return 0
}

func parseArgs(args []string) (*Config, error) {
	cfg := &Config{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			cfg.Help = true
		case "--dump", "-d":
			cfg.DumpMode = true
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

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "env-audit [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --file, -f <path>     Path to .env file to scan (optional)")
	fmt.Fprintln(w, "  --required, -r <vars> Comma-separated list of required variables")
	fmt.Fprintln(w, "  --dump, -d            Output parsed configuration (with redaction)")
	fmt.Fprintln(w, "  --help, -h            Show this help message")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Exit Codes:")
	fmt.Fprintln(w, "  0  No risks found")
	fmt.Fprintln(w, "  1  Risks detected")
	fmt.Fprintln(w, "  2  Fatal error (invalid arguments, file not found)")
}
