package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit-v2, Property 12: Config file parsing**
// **Validates: Requirements 10.1, 10.2**
// For any valid YAML config with supported fields, LoadFile SHALL parse it
// correctly and return a FileConfig with matching values.
func TestProperty_ConfigFileParsing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid file paths
	genFilePath := gen.AlphaString().Map(func(s string) string {
		if s == "" {
			return "test.env"
		}
		return s + ".env"
	})

	// Generator for list of required vars
	genRequired := gen.SliceOf(gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	}))

	// Generator for boolean values
	genBool := gen.Bool()

	properties.Property("valid YAML config is parsed correctly", prop.ForAll(
		func(file string, required []string, strict, checkLeaks, quiet, json, github bool) bool {
			// Create temp dir and config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".env-audit.yaml")

			// Build YAML content
			var lines []string
			if file != "" {
				lines = append(lines, "file: "+file)
			}
			if len(required) > 0 {
				lines = append(lines, "required:")
				for _, r := range required {
					lines = append(lines, "  - "+r)
				}
			}
			if strict {
				lines = append(lines, "strict: true")
			}
			if checkLeaks {
				lines = append(lines, "check_leaks: true")
			}
			if quiet {
				lines = append(lines, "quiet: true")
			}
			if json {
				lines = append(lines, "json: true")
			}
			if github {
				lines = append(lines, "github: true")
			}

			content := strings.Join(lines, "\n")
			if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
				return false
			}

			// Load and verify
			cfg, err := LoadFile(configPath)
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Verify all fields match
			if cfg.File != file {
				t.Logf("File mismatch: expected %q, got %q", file, cfg.File)
				return false
			}
			if len(cfg.Required) != len(required) {
				t.Logf("Required length mismatch: expected %d, got %d", len(required), len(cfg.Required))
				return false
			}
			for i, r := range required {
				if cfg.Required[i] != r {
					t.Logf("Required[%d] mismatch: expected %q, got %q", i, r, cfg.Required[i])
					return false
				}
			}
			if cfg.Strict != strict {
				t.Logf("Strict mismatch: expected %v, got %v", strict, cfg.Strict)
				return false
			}
			if cfg.CheckLeaks != checkLeaks {
				t.Logf("CheckLeaks mismatch: expected %v, got %v", checkLeaks, cfg.CheckLeaks)
				return false
			}
			if cfg.Quiet != quiet {
				t.Logf("Quiet mismatch: expected %v, got %v", quiet, cfg.Quiet)
				return false
			}
			if cfg.JSON != json {
				t.Logf("JSON mismatch: expected %v, got %v", json, cfg.JSON)
				return false
			}
			if cfg.GitHub != github {
				t.Logf("GitHub mismatch: expected %v, got %v", github, cfg.GitHub)
				return false
			}

			return true
		},
		genFilePath,
		genRequired,
		genBool,
		genBool,
		genBool,
		genBool,
		genBool,
	))

	properties.TestingRun(t)
}

// Unit tests for config loading
func TestLoadFile_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `file: .env
required:
  - DATABASE_URL
  - API_KEY
strict: true
check_leaks: true
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.File != ".env" {
		t.Errorf("expected file=.env, got %q", cfg.File)
	}
	if len(cfg.Required) != 2 {
		t.Errorf("expected 2 required vars, got %d", len(cfg.Required))
	}
	if !cfg.Strict {
		t.Error("expected strict=true")
	}
	if !cfg.CheckLeaks {
		t.Error("expected check_leaks=true")
	}
}

func TestLoadFile_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load empty config: %v", err)
	}

	if cfg.File != "" || len(cfg.Required) != 0 || cfg.Strict || cfg.CheckLeaks {
		t.Error("empty config should have zero values")
	}
}

func TestLoadFile_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `file: .env
required: [
  - this is invalid
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(configPath)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestLoadFile_FileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFindConfigFile_Found(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create .env-audit.yaml
	if err := os.WriteFile(".env-audit.yaml", []byte("file: test.env"), 0644); err != nil {
		t.Fatal(err)
	}

	found := FindConfigFile()
	if found != ".env-audit.yaml" {
		t.Errorf("expected .env-audit.yaml, got %q", found)
	}
}

func TestFindConfigFile_YmlExtension(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create .env-audit.yml
	if err := os.WriteFile(".env-audit.yml", []byte("file: test.env"), 0644); err != nil {
		t.Fatal(err)
	}

	found := FindConfigFile()
	if found != ".env-audit.yml" {
		t.Errorf("expected .env-audit.yml, got %q", found)
	}
}

func TestFindConfigFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	found := FindConfigFile()
	if found != "" {
		t.Errorf("expected empty string when no config, got %q", found)
	}
}

func TestFindConfigFile_PriorityYamlOverYml(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create both files
	if err := os.WriteFile(".env-audit.yaml", []byte("file: yaml.env"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".env-audit.yml", []byte("file: yml.env"), 0644); err != nil {
		t.Fatal(err)
	}

	found := FindConfigFile()
	if found != ".env-audit.yaml" {
		t.Errorf("expected .env-audit.yaml (higher priority), got %q", found)
	}
}

func TestFindConfigFileInDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config in subdir
	subdir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subdir, 0755)
	if err := os.WriteFile(filepath.Join(subdir, ".env-audit.yaml"), []byte("file: test.env"), 0644); err != nil {
		t.Fatal(err)
	}

	found := FindConfigFileInDir(subdir)
	expected := filepath.Join(subdir, ".env-audit.yaml")
	if found != expected {
		t.Errorf("expected %q, got %q", expected, found)
	}
}

