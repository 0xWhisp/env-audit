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
	// TODO: Implement in task 7
	os.Exit(0)
}

func parseArgs(args []string) (*Config, error) {
	// TODO: Implement in task 7
	return &Config{}, nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "env-audit [options]")
}
