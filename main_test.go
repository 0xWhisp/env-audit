package main

import (
	"bytes"
	"os"
	"path/filepath"
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
