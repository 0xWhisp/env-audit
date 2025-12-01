package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"env-audit/internal/audit"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit, Property 7: Exit code correctness**
// **Validates: Requirements 4.2, 4.3, 5.1, 5.2**
// For any ScanResult, if HasRisks is true the exit code SHALL be 1,
// if HasRisks is false the exit code SHALL be 0.
// Note: Empty values are warnings, not errors. They only cause exit 1 in strict mode.
func TestProperty_ExitCodeCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for environment entries that will produce errors (missing required)
	genEnvWithErrors := gen.MapOf(
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && !audit.IsSensitiveKey(s)
		}),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	)

	// Generator for environment entries without issues (non-empty, non-sensitive)
	genEnvWithoutIssues := gen.MapOf(
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && !audit.IsSensitiveKey(s)
		}),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	)

	// Property: When errors exist (missing required), exit code is 1
	properties.Property("exit code is 1 when risks detected", prop.ForAll(
		func(env map[string]string) bool {
			// Create temp .env file
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			content := ""
			for key, value := range env {
				content += key + "=" + value + "\n"
			}
			if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
				return false
			}

			// Use --required with a key that doesn't exist to create an error
			var stdout, stderr bytes.Buffer
			exitCode := Run([]string{"-f", envFile, "-r", "MISSING_REQUIRED_VAR"}, &stdout, &stderr)
			return exitCode == 1
		},
		genEnvWithErrors,
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
			exitCode := Run([]string{"-f", envFile}, &stdout, &stderr)
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
			exitCode := Run(args, &stdout, &stderr)
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
			exitCode := Run([]string{"-f", path}, &stdout, &stderr)
			return exitCode == 2
		},
		genMissingFilePath,
	))

	properties.TestingRun(t)
}


func TestRun_HelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := Run([]string{"-h"}, &stdout, &bytes.Buffer{})

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
	exitCode := Run([]string{"-f", tmpfile.Name(), "-d"}, &stdout, &stderr)

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
	exitCode := Run([]string{}, &stdout, &stderr)
	if exitCode == 2 {
		t.Error("should not be fatal error")
	}
}

// **Feature: env-audit-v2, Property 2: Quiet mode behavior**
// **Validates: Requirements 3.1, 3.2**
// For any audit scenario with --quiet flag, stdout SHALL be empty AND exit code
// SHALL correctly reflect issue presence (0=none, 1=issues, 2=error).
// Note: Empty values are warnings. Use --strict to make them cause exit 1.
func TestProperty_QuietModeBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid env key (alphanumeric, non-empty)
	genKey := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && !audit.IsSensitiveKey(s)
	})

	// Generator for non-empty value
	genNonEmptyValue := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})

	// Generator for env map with no issues (non-empty values, non-sensitive keys)
	genEnvNoIssues := gen.MapOf(genKey, genNonEmptyValue)

	// Generator for env map with issues (empty values)
	genEnvWithIssues := gen.MapOf(genKey, gen.Const("")).SuchThat(func(m map[string]string) bool {
		return len(m) > 0
	})

	// Property: Quiet mode with no issues -> stdout empty, exit code 0
	properties.Property("quiet mode: no issues -> empty stdout, exit 0", prop.ForAll(
		func(env map[string]string) bool {
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
			exitCode := Run([]string{"-f", envFile, "-q"}, &stdout, &stderr)

			// stdout must be empty
			if stdout.Len() != 0 {
				t.Logf("stdout not empty: %s", stdout.String())
				return false
			}
			// exit code must be 0
			if exitCode != 0 {
				t.Logf("exit code not 0: %d", exitCode)
				return false
			}
			return true
		},
		genEnvNoIssues,
	))

	// Property: Quiet mode with issues (strict mode) -> stdout empty, exit code 1
	// Empty values are warnings, so we use --strict to make them cause exit 1
	properties.Property("quiet mode: issues -> empty stdout, exit 1", prop.ForAll(
		func(env map[string]string) bool {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			content := ""
			for key := range env {
				content += key + "=\n" // Empty value creates warning
			}
			if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
				return false
			}

			var stdout, stderr bytes.Buffer
			// Use --strict to make warnings cause exit 1
			exitCode := Run([]string{"-f", envFile, "-q", "--strict"}, &stdout, &stderr)

			// stdout must be empty
			if stdout.Len() != 0 {
				t.Logf("stdout not empty: %s", stdout.String())
				return false
			}
			// exit code must be 1
			if exitCode != 1 {
				t.Logf("exit code not 1: %d", exitCode)
				return false
			}
			return true
		},
		genEnvWithIssues,
	))

	// Property: Quiet mode with fatal error -> stdout empty, exit code 2
	properties.Property("quiet mode: fatal error -> empty stdout, exit 2", prop.ForAll(
		func(filename string) bool {
			// Use non-existent file path
			path := "/nonexistent/path/" + filename + ".env"

			var stdout, stderr bytes.Buffer
			exitCode := Run([]string{"-f", path, "-q"}, &stdout, &stderr)

			// stdout must be empty
			if stdout.Len() != 0 {
				t.Logf("stdout not empty: %s", stdout.String())
				return false
			}
			// exit code must be 2
			if exitCode != 2 {
				t.Logf("exit code not 2: %d", exitCode)
				return false
			}
			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

func TestRun_QuietMode_SuppressesStdout(t *testing.T) {
	tmpfile, _ := os.CreateTemp("", "test*.env")
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("APP=test\n")
	tmpfile.Close()

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", tmpfile.Name(), "-q"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Errorf("quiet mode should suppress stdout, got: %s", stdout.String())
	}
}

func TestRun_QuietMode_MaintainsExitCode1(t *testing.T) {
	tmpfile, _ := os.CreateTemp("", "test*.env")
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("EMPTY_VAR=\n") // Empty value creates warning
	tmpfile.Close()

	var stdout, stderr bytes.Buffer
	// Use --strict to make warnings cause exit 1
	exitCode := Run([]string{"-f", tmpfile.Name(), "-q", "--strict"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("expected exit 1 for issues in strict mode, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Errorf("quiet mode should suppress stdout, got: %s", stdout.String())
	}
}

func TestRun_QuietMode_ErrorsToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", "/nonexistent/file.env", "-q"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("expected exit 2 for error, got %d", exitCode)
	}
	if stderr.Len() == 0 {
		t.Error("errors should still go to stderr in quiet mode")
	}
	if stdout.Len() != 0 {
		t.Errorf("quiet mode should suppress stdout, got: %s", stdout.String())
	}
}

func TestRun_QuietMode_DumpMode(t *testing.T) {
	tmpfile, _ := os.CreateTemp("", "test*.env")
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("APP=test\n")
	tmpfile.Close()

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", tmpfile.Name(), "-d", "-q"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Errorf("quiet mode should suppress dump output, got: %s", stdout.String())
	}
}

// **Feature: env-audit-v2, Property 4: Strict mode escalation**
// **Validates: Requirements 4.1, 4.2**
// For any environment with warning-level issues (empty values), when --strict flag
// is used, exit code SHALL be 1. Without strict, warnings don't cause exit 1.
func TestProperty_StrictModeEscalation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid env key (alphanumeric, non-empty, non-sensitive)
	genKey := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && !audit.IsSensitiveKey(s)
	})

	// Generator for env map with empty values (warnings)
	genEnvWithEmptyValues := gen.MapOf(genKey, gen.Const("")).SuchThat(func(m map[string]string) bool {
		return len(m) > 0
	})

	// Property: With --strict flag, empty values (warnings) cause exit code 1
	properties.Property("strict mode: warnings cause exit 1", prop.ForAll(
		func(env map[string]string) bool {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			content := ""
			for key := range env {
				content += key + "=\n" // Empty value creates warning
			}
			if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
				return false
			}

			var stdout, stderr bytes.Buffer
			exitCode := Run([]string{"-f", envFile, "--strict"}, &stdout, &stderr)

			// In strict mode, warnings should cause exit code 1
			return exitCode == 1
		},
		genEnvWithEmptyValues,
	))

	// Property: Without --strict flag, empty values (warnings) don't cause exit code 1
	properties.Property("non-strict mode: warnings don't cause exit 1", prop.ForAll(
		func(env map[string]string) bool {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			content := ""
			for key := range env {
				content += key + "=\n" // Empty value creates warning
			}
			if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
				return false
			}

			var stdout, stderr bytes.Buffer
			exitCode := Run([]string{"-f", envFile}, &stdout, &stderr)

			// Without strict mode, warnings should NOT cause exit code 1
			return exitCode == 0
		},
		genEnvWithEmptyValues,
	))

	properties.TestingRun(t)
}

// **Feature: env-audit-v2, Property 3: Quiet mode error output**
// **Validates: Requirements 3.3**
// For any error condition with --quiet flag, stderr SHALL contain the error message.
func TestProperty_QuietModeErrorOutput(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for non-existent file paths (file not found error)
	genMissingFilePath := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	}).Map(func(s string) string {
		return "/nonexistent/path/" + s + ".env"
	})

	// Property: Missing file with quiet mode -> stderr contains error
	properties.Property("quiet mode: missing file -> stderr contains error", prop.ForAll(
		func(path string) bool {
			var stdout, stderr bytes.Buffer
			exitCode := Run([]string{"-f", path, "-q"}, &stdout, &stderr)

			// Exit code must be 2 (fatal error)
			if exitCode != 2 {
				t.Logf("exit code not 2: %d", exitCode)
				return false
			}
			// stdout must be empty
			if stdout.Len() != 0 {
				t.Logf("stdout not empty: %s", stdout.String())
				return false
			}
			// stderr must contain error message
			if stderr.Len() == 0 {
				t.Logf("stderr is empty, should contain error")
				return false
			}
			return true
		},
		genMissingFilePath,
	))

	// Generator for invalid argument patterns that cause errors
	// Note: -q must come BEFORE flags that expect values, otherwise -q becomes the value
	genInvalidArgs := gen.OneGenOf(
		// Missing value for --file (at end of args)
		gen.Const([]string{"-q", "-f"}),
		// Missing value for --required (at end of args)
		gen.Const([]string{"-q", "-r"}),
		// Unknown flag (with quiet)
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }).
			Map(func(s string) []string { return []string{"-q", "--" + s + "invalid"} }),
	)

	// Property: Invalid args with quiet mode -> stderr contains error
	properties.Property("quiet mode: invalid args -> stderr contains error", prop.ForAll(
		func(args []string) bool {
			var stdout, stderr bytes.Buffer
			exitCode := Run(args, &stdout, &stderr)

			// Exit code must be 2 (fatal error)
			if exitCode != 2 {
				t.Logf("exit code not 2: %d", exitCode)
				return false
			}
			// stdout must be empty
			if stdout.Len() != 0 {
				t.Logf("stdout not empty: %s", stdout.String())
				return false
			}
			// stderr must contain error message
			if stderr.Len() == 0 {
				t.Logf("stderr is empty, should contain error")
				return false
			}
			return true
		},
		genInvalidArgs,
	))

	properties.TestingRun(t)
}
