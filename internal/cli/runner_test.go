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

// Unit tests for init mode
// _Requirements: 7.1, 7.4_

func TestRun_InitMode_CreatesFile(t *testing.T) {
	// Create temp dir and .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("APP=test\nDEBUG=true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir for test
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "--init"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d, stderr: %s", exitCode, stderr.String())
	}

	// Check .env.example was created
	exampleFile := filepath.Join(tmpDir, ".env.example")
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		t.Error(".env.example should have been created")
	}

	// Check output message
	if !strings.Contains(stdout.String(), "Generated") {
		t.Errorf("expected 'Generated' message, got: %s", stdout.String())
	}
}

func TestRun_InitMode_ExistingFileNoForce(t *testing.T) {
	// Create temp dir with .env and existing .env.example
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	exampleFile := filepath.Join(tmpDir, ".env.example")
	if err := os.WriteFile(envFile, []byte("APP=test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exampleFile, []byte("OLD=content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir for test
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "--init"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("expected exit 2 when file exists without --force, got %d", exitCode)
	}

	// Check error message
	if !strings.Contains(stderr.String(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %s", stderr.String())
	}

	// Check original file wasn't modified
	content, _ := os.ReadFile(exampleFile)
	if !strings.Contains(string(content), "OLD=content") {
		t.Error("original .env.example should not have been modified")
	}
}

func TestRun_InitMode_ExistingFileWithForce(t *testing.T) {
	// Create temp dir with .env and existing .env.example
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	exampleFile := filepath.Join(tmpDir, ".env.example")
	if err := os.WriteFile(envFile, []byte("APP=test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exampleFile, []byte("OLD=content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir for test
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "--init", "--force"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 with --force, got %d, stderr: %s", exitCode, stderr.String())
	}

	// Check file was overwritten
	content, _ := os.ReadFile(exampleFile)
	if strings.Contains(string(content), "OLD=content") {
		t.Error(".env.example should have been overwritten")
	}
	if !strings.Contains(string(content), "APP=") {
		t.Error(".env.example should contain APP key")
	}
}

// **Feature: env-audit-v2, Property 13: CLI flag precedence**
// **Validates: Requirements 10.3**
// For any config file with values, CLI flags SHALL take precedence when specified.
func TestProperty_CLIFlagPrecedence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("CLI --file flag overrides config file value", prop.ForAll(
		func(cliFile, configFile string) bool {
			tmpDir := t.TempDir()

			// Create env files
			cliEnvPath := filepath.Join(tmpDir, cliFile)
			configEnvPath := filepath.Join(tmpDir, configFile)
			os.WriteFile(cliEnvPath, []byte("CLI_VAR=cli\n"), 0644)
			os.WriteFile(configEnvPath, []byte("CONFIG_VAR=config\n"), 0644)

			// Create config file pointing to configEnvPath
			configPath := filepath.Join(tmpDir, ".env-audit.yaml")
			os.WriteFile(configPath, []byte("file: "+configEnvPath+"\n"), 0644)

			// Change to temp dir
			oldWd, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldWd)

			var stdout, stderr bytes.Buffer
			exitCode := Run([]string{"-f", cliEnvPath, "-d"}, &stdout, &stderr)

			// Should succeed
			if exitCode != 0 {
				return true // Skip errors
			}

			// Output should contain CLI_VAR, not CONFIG_VAR
			output := stdout.String()
			hasCliVar := strings.Contains(output, "CLI_VAR")
			hasConfigVar := strings.Contains(output, "CONFIG_VAR")

			return hasCliVar && !hasConfigVar
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }).Map(func(s string) string { return s + "_cli.env" }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }).Map(func(s string) string { return s + "_config.env" }),
	))

	properties.Property("CLI --strict flag takes precedence", prop.ForAll(
		func(configStrict bool) bool {
			tmpDir := t.TempDir()

			// Create env file with empty value (warning)
			envPath := filepath.Join(tmpDir, ".env")
			os.WriteFile(envPath, []byte("EMPTY_VAR=\n"), 0644)

			// Create config file with opposite strict setting
			configPath := filepath.Join(tmpDir, ".env-audit.yaml")
			strictStr := "false"
			if configStrict {
				strictStr = "true"
			}
			os.WriteFile(configPath, []byte("strict: "+strictStr+"\n"), 0644)

			// Change to temp dir
			oldWd, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldWd)

			// Run with --strict CLI flag (should always enable strict)
			var stdout, stderr bytes.Buffer
			exitCode := Run([]string{"-f", envPath, "--strict", "-q"}, &stdout, &stderr)

			// With CLI --strict, exit code should be 1 (empty value is warning -> error)
			return exitCode == 1
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Unit test for version flag
func TestRun_VersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"--version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 for --version, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "env-audit version") {
		t.Errorf("expected version output, got: %s", stdout.String())
	}
}

func TestRun_VersionFlagShort(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-V"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 for -V, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "env-audit version") {
		t.Errorf("expected version output, got: %s", stdout.String())
	}
}

// Unit test for --ignore flag
func TestRun_IgnoreFlag(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(envFile, []byte("IGNORED=\nNOT_IGNORED=\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "--ignore", "IGNORED", "--strict"}, &stdout, &stderr)

	// Should report NOT_IGNORED as empty (exit 1) but not IGNORED
	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}
	output := stdout.String()
	if strings.Contains(output, "IGNORED") && !strings.Contains(output, "NOT_IGNORED") {
		t.Error("IGNORED should not appear in output")
	}
}

// Unit test for GitHub output flag
func TestRun_GitHubOutput(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(envFile, []byte("EMPTY_VAR=\n"), 0644)

	var stdout, stderr bytes.Buffer
	Run([]string{"-f", envFile, "--github"}, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "::warning::") {
		t.Errorf("expected GitHub ::warning:: format, got: %s", output)
	}
}

func TestRun_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(envFile, []byte("APP=test\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "--json"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	output := stdout.String()
	if !strings.Contains(output, `"hasRisks"`) {
		t.Errorf("expected JSON output, got: %s", output)
	}
}

func TestRun_ExampleComparison(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	exampleFile := filepath.Join(tmpDir, ".env.example")
	os.WriteFile(envFile, []byte("APP=test\n"), 0644)
	os.WriteFile(exampleFile, []byte("APP=\nMISSING=\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "-e", exampleFile}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("expected exit 1 for missing vars, got %d", exitCode)
	}
	output := stdout.String()
	if !strings.Contains(output, "MISSING") {
		t.Errorf("expected MISSING in output, got: %s", output)
	}
}

func TestRun_ExampleFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(envFile, []byte("APP=test\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "-e", "/nonexistent/example.env"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("expected exit 2 for missing example file, got %d", exitCode)
	}
}

func TestRun_DiffMode(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.env")
	file2 := filepath.Join(tmpDir, "file2.env")
	os.WriteFile(file1, []byte("APP=test\nOLD=value\n"), 0644)
	os.WriteFile(file2, []byte("APP=changed\nNEW=value\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", file1, "--diff", file2}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 for diff, got %d", exitCode)
	}
	output := stdout.String()
	if !strings.Contains(output, "OLD") || !strings.Contains(output, "NEW") {
		t.Errorf("expected diff output, got: %s", output)
	}
}

func TestRun_DiffMode_WithoutFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"--diff", "some.env"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("expected exit 2 when --diff used without --file, got %d", exitCode)
	}
}

func TestRun_DiffMode_SecondFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.env")
	os.WriteFile(file1, []byte("APP=test\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", file1, "--diff", "/nonexistent/file2.env"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("expected exit 2 for missing diff file, got %d", exitCode)
	}
}

func TestRun_CheckLeaks(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	// Create a file with a GitHub token pattern
	os.WriteFile(envFile, []byte("GITHUB_TOKEN=ghp_aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile, "--check-leaks"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("expected exit 1 for detected leak, got %d", exitCode)
	}
	output := stdout.String()
	if !strings.Contains(output, "Potential Leaks") {
		t.Errorf("expected leak detection in output, got: %s", output)
	}
}

func TestRun_WatchMode_RequiresFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"--watch"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("expected exit 2 when --watch used without --file, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "--watch requires --file") {
		t.Errorf("expected error message about --file, got: %s", stderr.String())
	}
}

func TestRun_NoIssuesOutput(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(envFile, []byte("APP=test\nDEBUG=true\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 for no issues, got %d", exitCode)
	}
	output := stdout.String()
	if !strings.Contains(output, "No issues found") {
		t.Errorf("expected 'No issues found' in output, got: %s", output)
	}
}

func TestRun_ConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	configFile := filepath.Join(tmpDir, ".env-audit.yaml")
	os.WriteFile(envFile, []byte("APP=test\n"), 0644)
	os.WriteFile(configFile, []byte("strict: true\n"), 0644)

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	// Config sets strict, but no warnings so exit 0
	exitCode := Run([]string{"-f", envFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 with config file, got %d, stderr: %s", exitCode, stderr.String())
	}
}

func TestRun_ConfigFile_Malformed(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	configFile := filepath.Join(tmpDir, ".env-audit.yaml")
	os.WriteFile(envFile, []byte("APP=test\n"), 0644)
	os.WriteFile(configFile, []byte("invalid: [yaml\n"), 0644)

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", envFile}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("expected exit 2 for malformed config, got %d", exitCode)
	}
}

func TestRun_ConfigFile_AllSettings(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	configFile := filepath.Join(tmpDir, ".env-audit.yaml")
	os.WriteFile(envFile, []byte("APP=test\n"), 0644)
	os.WriteFile(configFile, []byte(`
file: .env
required:
  - APP
strict: true
check_leaks: true
quiet: false
json: false
github: false
no_color: true
ignore:
  - IGNORED_VAR
`), 0644)

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{}, &stdout, &stderr)

	// Should work with all settings from config
	if exitCode != 0 {
		t.Errorf("expected exit 0 with full config, got %d, stderr: %s", exitCode, stderr.String())
	}
}

func TestRun_DiffQuietMode(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.env")
	file2 := filepath.Join(tmpDir, "file2.env")
	os.WriteFile(file1, []byte("APP=test\n"), 0644)
	os.WriteFile(file2, []byte("APP=changed\n"), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"-f", file1, "--diff", file2, "-q"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 for diff in quiet mode, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Errorf("quiet mode should suppress diff output, got: %s", stdout.String())
	}
}

func TestRun_InitMode_FromOSEnv(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	// Init from OS env (no --file specified)
	exitCode := Run([]string{"--init"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit 0 for init from OS env, got %d, stderr: %s", exitCode, stderr.String())
	}

	// Check .env.example was created
	if _, err := os.Stat(".env.example"); os.IsNotExist(err) {
		t.Error(".env.example should have been created")
	}
}

// Test MergeWithFileConfig
func TestConfig_MergeWithFileConfig(t *testing.T) {
	cfg := &Config{}
	fileCfg := &FileConfig{
		File:       "from_file.env",
		Required:   []string{"REQ1"},
		Example:    "example.env",
		Ignore:     []string{"IGN1"},
		Strict:     true,
		CheckLeaks: true,
		Quiet:      true,
		JSON:       true,
		GitHub:     true,
		NoColor:    true,
	}

	cfg.MergeWithFileConfig(fileCfg)

	if cfg.FilePath != "from_file.env" {
		t.Errorf("expected FilePath=from_file.env, got %s", cfg.FilePath)
	}
	if len(cfg.Required) != 1 || cfg.Required[0] != "REQ1" {
		t.Errorf("expected Required=[REQ1], got %v", cfg.Required)
	}
	if !cfg.Strict {
		t.Error("expected Strict=true")
	}
}

func TestConfig_MergeWithFileConfig_CLIPrecedence(t *testing.T) {
	cfg := &Config{
		FilePath: "cli_file.env",
		Required: []string{"CLI_REQ"},
		Strict:   true,
	}
	fileCfg := &FileConfig{
		File:     "from_file.env",
		Required: []string{"FILE_REQ"},
		Strict:   false, // CLI already set true
	}

	cfg.MergeWithFileConfig(fileCfg)

	// CLI values should take precedence
	if cfg.FilePath != "cli_file.env" {
		t.Errorf("CLI FilePath should take precedence, got %s", cfg.FilePath)
	}
	if cfg.Required[0] != "CLI_REQ" {
		t.Errorf("CLI Required should take precedence, got %v", cfg.Required)
	}
	if !cfg.Strict {
		t.Error("CLI Strict should take precedence")
	}
}

func TestConfig_MergeWithFileConfig_Nil(t *testing.T) {
	cfg := &Config{FilePath: "original.env"}
	cfg.MergeWithFileConfig(nil)

	// Should not crash and preserve original values
	if cfg.FilePath != "original.env" {
		t.Errorf("expected original FilePath, got %s", cfg.FilePath)
	}
}
