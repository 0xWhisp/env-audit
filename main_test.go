package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit, Property 7: Exit code correctness**
// **Validates: Requirements 4.2, 4.3, 5.1, 5.2**
// For any ScanResult, if HasRisks is true the exit code SHALL be 1,
// if HasRisks is false the exit code SHALL be 0.
func TestProperty_ExitCodeCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for environment entries that will produce issues (empty values)
	genEnvWithIssues := gen.MapOf(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.Const(""), // Empty values create issues
	).SuchThat(func(m map[string]string) bool { return len(m) > 0 })

	// Generator for environment entries without issues (non-empty, non-sensitive)
	genEnvWithoutIssues := gen.MapOf(
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && !IsSensitiveKey(s)
		}),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	)

	// Property: When issues exist, exit code is 1
	properties.Property("exit code is 1 when risks detected", prop.ForAll(
		func(env map[string]string) bool {
			// Create temp .env file with empty values
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			content := ""
			for key := range env {
				content += key + "=\n" // Empty value creates issue
			}
			if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
				return false
			}

			var stdout, stderr bytes.Buffer
			exitCode := run([]string{"-f", envFile}, &stdout, &stderr)
			return exitCode == 1
		},
		genEnvWithIssues,
	))

	// Property: When no issues exist, exit code is 0
	properties.Property("exit code is 0 when no risks detected", prop.ForAll(
		func(env map[string]string) bool {
			// Create temp .env file with non-empty, non-sensitive values
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			content := ""
			for key, value := range env {
				content += key + "=" + value + "\n"
			}
			if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
				return false
			}

			var stdout, stderr bytes.Buffer
			exitCode := run([]string{"-f", envFile}, &stdout, &stderr)
			return exitCode == 0
		},
		genEnvWithoutIssues,
	))

	properties.TestingRun(t)
}

// **Feature: env-audit, Property 8: Fatal error exit code**
// **Validates: Requirements 5.3**
// For any fatal error condition (invalid arguments, missing file),
// the exit code SHALL be 2.
func TestProperty_FatalErrorExitCode(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for invalid argument patterns
	genInvalidArgs := gen.OneGenOf(
		// Unknown argument
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }).
			Map(func(s string) []string { return []string{"--" + s + "invalid"} }),
		// Missing value for --file
		gen.Const([]string{"-f"}),
		// Missing value for --required
		gen.Const([]string{"-r"}),
		// Missing value for --file at end
		gen.Const([]string{"--file"}),
		// Missing value for --required at end
		gen.Const([]string{"--required"}),
	)

	// Property: Invalid arguments produce exit code 2
	properties.Property("invalid arguments produce exit code 2", prop.ForAll(
		func(args []string) bool {
			var stdout, stderr bytes.Buffer
			exitCode := run(args, &stdout, &stderr)
			return exitCode == 2
		},
		genInvalidArgs,
	))

	// Generator for non-existent file paths
	genMissingFilePath := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	}).Map(func(s string) string {
		return "/nonexistent/path/" + s + ".env"
	})

	// Property: Missing file produces exit code 2
	properties.Property("missing file produces exit code 2", prop.ForAll(
		func(path string) bool {
			var stdout, stderr bytes.Buffer
			exitCode := run([]string{"-f", path}, &stdout, &stderr)
			return exitCode == 2
		},
		genMissingFilePath,
	))

	properties.TestingRun(t)
}


// Unit tests for CLI argument parsing
// _Requirements: 4.4, 5.3_

func TestParseArgs_ValidArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name:     "help flag long",
			args:     []string{"--help"},
			expected: Config{Help: true},
		},
		{
			name:     "help flag short",
			args:     []string{"-h"},
			expected: Config{Help: true},
		},
		{
			name:     "dump flag long",
			args:     []string{"--dump"},
			expected: Config{DumpMode: true},
		},
		{
			name:     "dump flag short",
			args:     []string{"-d"},
			expected: Config{DumpMode: true},
		},
		{
			name:     "file flag long",
			args:     []string{"--file", ".env"},
			expected: Config{FilePath: ".env"},
		},
		{
			name:     "file flag short",
			args:     []string{"-f", "config.env"},
			expected: Config{FilePath: "config.env"},
		},
		{
			name:     "required flag long",
			args:     []string{"--required", "VAR1,VAR2"},
			expected: Config{Required: []string{"VAR1", "VAR2"}},
		},
		{
			name:     "required flag short",
			args:     []string{"-r", "API_KEY"},
			expected: Config{Required: []string{"API_KEY"}},
		},
		{
			name:     "multiple flags combined",
			args:     []string{"-f", "test.env", "-r", "A,B", "-d"},
			expected: Config{FilePath: "test.env", Required: []string{"A", "B"}, DumpMode: true},
		},
		{
			name:     "no args",
			args:     []string{},
			expected: Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Help != tt.expected.Help {
				t.Errorf("Help: got %v, want %v", cfg.Help, tt.expected.Help)
			}
			if cfg.DumpMode != tt.expected.DumpMode {
				t.Errorf("DumpMode: got %v, want %v", cfg.DumpMode, tt.expected.DumpMode)
			}
			if cfg.FilePath != tt.expected.FilePath {
				t.Errorf("FilePath: got %v, want %v", cfg.FilePath, tt.expected.FilePath)
			}
			if len(cfg.Required) != len(tt.expected.Required) {
				t.Errorf("Required length: got %v, want %v", len(cfg.Required), len(tt.expected.Required))
			}
			for i := range cfg.Required {
				if cfg.Required[i] != tt.expected.Required[i] {
					t.Errorf("Required[%d]: got %v, want %v", i, cfg.Required[i], tt.expected.Required[i])
				}
			}
		})
	}
}

func TestParseArgs_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"--unknown"}},
		{name: "missing file value", args: []string{"--file"}},
		{name: "missing file value short", args: []string{"-f"}},
		{name: "missing required value", args: []string{"--required"}},
		{name: "missing required value short", args: []string{"-r"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseArgs(tt.args)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseArgs_HelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := run([]string{"-h"}, &stdout, &bytes.Buffer{})

	if exitCode != 0 {
		t.Errorf("help flag exit code: got %d, want 0", exitCode)
	}
	if stdout.Len() == 0 {
		t.Error("help flag should produce output")
	}
}


func TestRun_DumpMode(t *testing.T) {
	tmpfile, _ := os.CreateTemp("", "test*.env")
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("APP=test\n")
	tmpfile.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-f", tmpfile.Name(), "-d"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "APP=test") {
		t.Error("dump should contain APP=test")
	}
}

func TestRun_NoFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	// Run without file flag uses os.Environ - just verify it doesn't crash
	exitCode := run([]string{}, &stdout, &stderr)
	if exitCode == 2 {
		t.Error("should not be fatal error")
	}
}

func TestParseCommaSeparated_Empty(t *testing.T) {
	result := parseCommaSeparated("")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestTrimSpace_AllSpaces(t *testing.T) {
	result := trimSpace("   ")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestTrimSpace_Tabs(t *testing.T) {
	result := trimSpace("\t\tvalue\t\t")
	if result != "value" {
		t.Errorf("expected 'value', got %q", result)
	}
}
