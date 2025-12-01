package cli

import (
	"fmt"
	"io"

	"env-audit/internal/audit"
	"env-audit/internal/parser"
)

// Run executes the main logic and returns the exit code
func Run(args []string, stdout, stderr io.Writer) int {
	cfg, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}

	if cfg.Help {
		PrintUsage(stdout)
		return 0
	}

	var env map[string]string
	var duplicates []string

	if cfg.FilePath != "" {
		result, err := parser.ParseEnvFile(cfg.FilePath)
		if err != nil {
			fmt.Fprintln(stderr, "Error:", err)
			return 2
		}
		env = result.Entries
		duplicates = result.Duplicates
	} else {
		env = parser.ReadOSEnv()
	}

	if cfg.DumpMode {
		fmt.Fprintln(stdout, parser.FormatEnv(env, true))
		return 0
	}

	scanResult := audit.Scan(env, &audit.ScanOptions{
		Required:   cfg.Required,
		Duplicates: duplicates,
	})
	fmt.Fprint(stdout, FormatSummary(scanResult))

	if scanResult.HasRisks {
		return 1
	}
	return 0
}
