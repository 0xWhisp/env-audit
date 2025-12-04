package cli

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"env-audit/internal/audit"
	"env-audit/internal/config"
	"env-audit/internal/parser"

	"github.com/fsnotify/fsnotify"
)

// Version is the current version of env-audit
const Version = "0.2.0"

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

	if cfg.Version {
		fmt.Fprintln(stdout, "env-audit version", Version)
		return 0
	}

	// Load and merge config file if present
	if configPath := config.FindConfigFile(); configPath != "" {
		fileCfg, err := config.LoadFile(configPath)
		if err != nil {
			fmt.Fprintln(stderr, "Error:", err)
			return 2
		}
		cfg.MergeWithFileConfig(&FileConfig{
			File:       fileCfg.File,
			Required:   fileCfg.Required,
			Example:    fileCfg.Example,
			Ignore:     fileCfg.Ignore,
			Strict:     fileCfg.Strict,
			CheckLeaks: fileCfg.CheckLeaks,
			Quiet:      fileCfg.Quiet,
			JSON:       fileCfg.JSON,
			GitHub:     fileCfg.GitHub,
			NoColor:    fileCfg.NoColor,
		})
	}

	// Handle watch mode - continuous file watching
	if cfg.Watch {
		return runWatch(cfg, stdout, stderr)
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

	// Handle init mode - generate .env.example
	if cfg.Init {
		return runInit(env, cfg.Force, stdout, stderr)
	}

	// Handle diff mode - compare two env files
	if cfg.DiffFile != "" {
		if cfg.FilePath == "" {
			fmt.Fprintln(stderr, "Error: --diff requires --file to specify the first file")
			return 2
		}
		return runDiff(cfg.FilePath, cfg.DiffFile, cfg.Quiet, stdout, stderr)
	}

	if cfg.DumpMode {
		if !cfg.Quiet {
			fmt.Fprintln(stdout, parser.FormatEnv(env, true))
		}
		return 0
	}

	// Handle example file comparison
	var missing, extra []string
	if cfg.ExampleFile != "" {
		exampleResult, err := parser.ParseEnvFile(cfg.ExampleFile)
		if err != nil {
			fmt.Fprintln(stderr, "Error:", err)
			return 2
		}
		compareResult := parser.Compare(env, exampleResult.Entries)
		missing = compareResult.Missing
		extra = compareResult.Extra
	}

	scanResult := audit.Scan(env, &audit.ScanOptions{
		Required:   cfg.Required,
		Ignore:     cfg.Ignore,
		Duplicates: duplicates,
		Missing:    missing,
		Extra:      extra,
		CheckLeaks: cfg.CheckLeaks,
		Strict:     cfg.Strict,
	})

	if !cfg.Quiet {
		var output string
		if cfg.JSONOutput {
			formatter := &JSONFormatter{}
			output = formatter.Format(scanResult)
		} else if cfg.GitHubOutput {
			formatter := &GitHubFormatter{}
			output = formatter.Format(scanResult)
		} else {
			output = FormatSummary(scanResult)
		}
		if output != "" {
			fmt.Fprintln(stdout, output)
		}
	}

	if scanResult.HasRisks {
		return 1
	}
	return 0
}

// runWatch starts file watching mode
func runWatch(cfg *Config, stdout, stderr io.Writer) int {
	if cfg.FilePath == "" {
		fmt.Fprintln(stderr, "Error: --watch requires --file to specify a file to watch")
		return 2
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}
	defer watcher.Close()

	if err := watcher.Add(cfg.FilePath); err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Fprintln(stdout, "Watching", cfg.FilePath, "for changes... (Ctrl+C to stop)")

	// Run initial audit
	runAudit(cfg, stdout, stderr)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return 0
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Fprintln(stdout, "\n--- File changed ---")
				runAudit(cfg, stdout, stderr)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return 0
			}
			fmt.Fprintln(stderr, "Error:", err)
		case <-sigChan:
			fmt.Fprintln(stdout, "\nStopping watch mode...")
			return 0
		}
	}
}

// runAudit performs a single audit run (used by watch mode)
func runAudit(cfg *Config, stdout, stderr io.Writer) int {
	result, err := parser.ParseEnvFile(cfg.FilePath)
	if err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}

	var missing, extra []string
	if cfg.ExampleFile != "" {
		exampleResult, err := parser.ParseEnvFile(cfg.ExampleFile)
		if err != nil {
			fmt.Fprintln(stderr, "Error:", err)
			return 2
		}
		compareResult := parser.Compare(result.Entries, exampleResult.Entries)
		missing = compareResult.Missing
		extra = compareResult.Extra
	}

	scanResult := audit.Scan(result.Entries, &audit.ScanOptions{
		Required:   cfg.Required,
		Ignore:     cfg.Ignore,
		Duplicates: result.Duplicates,
		Missing:    missing,
		Extra:      extra,
		CheckLeaks: cfg.CheckLeaks,
		Strict:     cfg.Strict,
	})

	if !cfg.Quiet {
		var output string
		if cfg.JSONOutput {
			formatter := &JSONFormatter{}
			output = formatter.Format(scanResult)
		} else if cfg.GitHubOutput {
			formatter := &GitHubFormatter{}
			output = formatter.Format(scanResult)
		} else {
			output = FormatSummary(scanResult)
		}
		if output != "" {
			fmt.Fprintln(stdout, output)
		}
	}

	if scanResult.HasRisks {
		return 1
	}
	return 0
}

// runInit generates a .env.example file from the current environment
func runInit(env map[string]string, force bool, stdout, stderr io.Writer) int {
	const outputFile = ".env.example"

	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil {
		if !force {
			fmt.Fprintln(stderr, "Error:", outputFile, "already exists (use --force to overwrite)")
			return 2
		}
	}

	template := parser.GenerateTemplate(env)
	if err := os.WriteFile(outputFile, []byte(template+"\n"), 0644); err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}

	fmt.Fprintln(stdout, "Generated", outputFile)
	return 0
}

// runDiff compares two env files and outputs the differences
func runDiff(file1, file2 string, quiet bool, stdout, stderr io.Writer) int {
	// Parse first file
	result1, err := parser.ParseEnvFile(file1)
	if err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}

	// Parse second file
	result2, err := parser.ParseEnvFile(file2)
	if err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}

	// Compute diff
	diffResult := parser.Diff(result1.Entries, result2.Entries)

	// Output diff (redact sensitive values)
	if !quiet {
		output := parser.FormatDiff(diffResult, true)
		if output != "" {
			fmt.Fprintln(stdout, output)
		}
	}

	return 0
}
