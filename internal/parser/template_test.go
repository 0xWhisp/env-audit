package parser

import (
	"strings"
	"testing"

	"env-audit/internal/audit"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit-v2, Property 8: Template generation completeness**
// **Validates: Requirements 7.2, 7.3**
// For any environment map, generated template SHALL contain all keys with values
// either empty or placeholder for non-sensitive keys, and SHALL NOT contain actual
// values for sensitive keys.
func TestProperty_TemplateGenerationCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid key names (non-empty alphanumeric with underscores)
	genKey := gen.RegexMatch(`[A-Z][A-Z0-9_]{0,19}`).SuchThat(func(s string) bool {
		return len(s) > 0
	})

	// Generator for values (non-empty to test redaction)
	genValue := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})

	// Generator for env maps
	genEnvMap := gen.MapOf(genKey, genValue).SuchThat(func(m map[string]string) bool {
		return len(m) > 0
	})

	properties.Property("Template contains all keys and redacts sensitive values", prop.ForAll(
		func(env map[string]string) bool {
			template := GenerateTemplate(env)
			lines := strings.Split(template, "\n")

			// Build map of template entries
			templateEntries := make(map[string]string)
			for _, line := range lines {
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					templateEntries[parts[0]] = parts[1]
				}
			}

			// Check all keys are present
			if len(templateEntries) != len(env) {
				return false
			}

			for key, originalValue := range env {
				templateValue, exists := templateEntries[key]
				if !exists {
					return false
				}

				if audit.IsSensitiveKey(key) {
					// Sensitive keys must have empty values
					if templateValue != "" {
						return false
					}
					// Must not contain original value
					if strings.Contains(template, originalValue) && originalValue != "" {
						return false
					}
				} else {
					// Non-sensitive keys must have placeholder (not original value)
					if templateValue == originalValue {
						return false
					}
					// Must have some placeholder value
					if templateValue == "" {
						return false
					}
				}
			}

			return true
		},
		genEnvMap,
	))

	properties.TestingRun(t)
}

func TestGenerateTemplate_Empty(t *testing.T) {
	result := GenerateTemplate(map[string]string{})
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGenerateTemplate_NonSensitiveKeys(t *testing.T) {
	env := map[string]string{
		"DATABASE_URL": "postgres://localhost/db",
		"PORT":         "3000",
	}
	result := GenerateTemplate(env)

	if !strings.Contains(result, "DATABASE_URL=your_database_url_here") {
		t.Errorf("expected placeholder for DATABASE_URL, got %q", result)
	}
	if !strings.Contains(result, "PORT=your_port_here") {
		t.Errorf("expected placeholder for PORT, got %q", result)
	}
}

func TestGenerateTemplate_SensitiveKeys(t *testing.T) {
	env := map[string]string{
		"API_KEY":     "secret123",
		"DB_PASSWORD": "pass456",
		"AUTH_TOKEN":  "token789",
	}
	result := GenerateTemplate(env)

	// Sensitive keys should have empty values
	if !strings.Contains(result, "API_KEY=") || strings.Contains(result, "API_KEY=your_") {
		t.Errorf("expected empty value for API_KEY, got %q", result)
	}
	if !strings.Contains(result, "DB_PASSWORD=") || strings.Contains(result, "DB_PASSWORD=your_") {
		t.Errorf("expected empty value for DB_PASSWORD, got %q", result)
	}
	if !strings.Contains(result, "AUTH_TOKEN=") || strings.Contains(result, "AUTH_TOKEN=your_") {
		t.Errorf("expected empty value for AUTH_TOKEN, got %q", result)
	}

	// Should not contain actual values
	if strings.Contains(result, "secret123") {
		t.Error("template should not contain actual secret value")
	}
}

func TestGenerateTemplate_MixedKeys(t *testing.T) {
	env := map[string]string{
		"APP_NAME":   "myapp",
		"SECRET_KEY": "supersecret",
		"DEBUG":      "true",
	}
	result := GenerateTemplate(env)

	// Non-sensitive should have placeholders
	if !strings.Contains(result, "APP_NAME=your_app_name_here") {
		t.Errorf("expected placeholder for APP_NAME, got %q", result)
	}
	if !strings.Contains(result, "DEBUG=your_debug_here") {
		t.Errorf("expected placeholder for DEBUG, got %q", result)
	}

	// Sensitive should be empty
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "SECRET_KEY=") {
			if line != "SECRET_KEY=" {
				t.Errorf("expected SECRET_KEY=, got %q", line)
			}
		}
	}
}

func TestGenerateTemplate_SortedOutput(t *testing.T) {
	env := map[string]string{
		"ZEBRA": "z",
		"APPLE": "a",
		"MANGO": "m",
	}
	result := GenerateTemplate(env)
	lines := strings.Split(result, "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "APPLE=") {
		t.Errorf("expected first line to be APPLE, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "MANGO=") {
		t.Errorf("expected second line to be MANGO, got %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "ZEBRA=") {
		t.Errorf("expected third line to be ZEBRA, got %q", lines[2])
	}
}
