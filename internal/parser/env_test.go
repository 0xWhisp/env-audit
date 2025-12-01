package parser

import (
	"os"
	"strings"
	"testing"

	"env-audit/internal/audit"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Unit tests for parser edge cases
// Requirements: 3.2, 3.3

func TestParseEnvFile_EmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	result, err := ParseEnvFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
	if len(result.Duplicates) != 0 {
		t.Errorf("expected 0 duplicates, got %d", len(result.Duplicates))
	}
}

func TestParseEnvFile_CommentsOnly(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "# This is a comment\n# Another comment\n  # Indented comment"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result, err := ParseEnvFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}


func TestParseEnvFile_MalformedLines(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "VALID=value\nmalformed_no_equals\nANOTHER=test"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result, err := ParseEnvFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.Entries["VALID"] != "value" {
		t.Errorf("expected VALID=value, got %s", result.Entries["VALID"])
	}
	if result.Entries["ANOTHER"] != "test" {
		t.Errorf("expected ANOTHER=test, got %s", result.Entries["ANOTHER"])
	}
}

func TestParseEnvFile_QuotedValues(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `DOUBLE="double quoted"
SINGLE='single quoted'
UNQUOTED=no quotes
EMPTY_DOUBLE=""
EMPTY_SINGLE=''`
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result, err := ParseEnvFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"DOUBLE", "double quoted"},
		{"SINGLE", "single quoted"},
		{"UNQUOTED", "no quotes"},
		{"EMPTY_DOUBLE", ""},
		{"EMPTY_SINGLE", ""},
	}

	for _, tc := range tests {
		if result.Entries[tc.key] != tc.expected {
			t.Errorf("key %s: expected %q, got %q", tc.key, tc.expected, result.Entries[tc.key])
		}
	}
}

func TestParseEnvFile_FileNotFound(t *testing.T) {
	_, err := ParseEnvFile("/nonexistent/path/file.env")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}


// **Feature: env-audit, Property 6: Duplicate key detection**
// **Validates: Requirements 3.4**
// For any .env content containing duplicate key definitions, ParseEnvFile SHALL
// include all duplicated key names in the Duplicates slice.
func TestProperty_DuplicateKeyDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid key names (alphanumeric + underscore)
	genKey := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})

	// Generator for values (any string without newlines)
	genValue := gen.AnyString().Map(func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", "")
	})

	properties.Property("detects all duplicate keys", prop.ForAll(
		func(keys []string, values []string) bool {
			if len(keys) == 0 || len(values) == 0 {
				return true
			}

			// Build .env content with intentional duplicates
			var lines []string
			duplicateCount := make(map[string]int)

			for i, key := range keys {
				value := values[i%len(values)]
				lines = append(lines, key+"="+value)
				duplicateCount[key]++
			}

			// Calculate expected duplicates
			expectedDuplicates := make(map[string]int)
			for key, count := range duplicateCount {
				if count > 1 {
					expectedDuplicates[key] = count - 1 // First occurrence is not a duplicate
				}
			}

			// Write to temp file
			tmpfile, err := os.CreateTemp("", "test*.env")
			if err != nil {
				return true // Skip on temp file error
			}
			defer os.Remove(tmpfile.Name())

			content := strings.Join(lines, "\n")
			if _, err := tmpfile.WriteString(content); err != nil {
				tmpfile.Close()
				return true
			}
			tmpfile.Close()

			// Parse and check
			result, err := ParseEnvFile(tmpfile.Name())
			if err != nil {
				return false
			}

			// Count duplicates found
			foundDuplicates := make(map[string]int)
			for _, key := range result.Duplicates {
				foundDuplicates[key]++
			}

			// Verify all expected duplicates are found
			for key, expectedCount := range expectedDuplicates {
				if foundDuplicates[key] != expectedCount {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(10, genKey),
		gen.SliceOfN(5, genValue),
	))

	properties.TestingRun(t)
}


// **Feature: env-audit, Property 5: .env parsing round-trip**
// **Validates: Requirements 3.2, 8.3**
// For any valid .env content (KEY=VALUE pairs without duplicates), parsing then
// formatting SHALL produce content that when parsed again yields the same key-value map.
func TestProperty_ParsingRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid key names (alphanumeric, no sensitive patterns)
	genSafeKey := gen.AlphaString().SuchThat(func(s string) bool {
		if len(s) == 0 {
			return false
		}
		// Exclude sensitive keys to avoid redaction affecting round-trip
		return !audit.IsSensitiveKey(s)
	})

	// Generator for safe values (no newlines, no quotes that would affect parsing)
	genSafeValue := gen.AlphaString()

	// Generator for unique key-value maps
	genEnvMap := gen.MapOf(genSafeKey, genSafeValue).SuchThat(func(m map[string]string) bool {
		return len(m) > 0
	})

	properties.Property("parse then format then parse yields same map", prop.ForAll(
		func(original map[string]string) bool {
			// Format the original map
			formatted := FormatEnv(original, false)

			// Write to temp file
			tmpfile, err := os.CreateTemp("", "test*.env")
			if err != nil {
				return true // Skip on temp file error
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.WriteString(formatted); err != nil {
				tmpfile.Close()
				return true
			}
			tmpfile.Close()

			// Parse the formatted content
			result, err := ParseEnvFile(tmpfile.Name())
			if err != nil {
				return false
			}

			// Compare maps
			if len(result.Entries) != len(original) {
				return false
			}

			for key, value := range original {
				if result.Entries[key] != value {
					return false
				}
			}

			return true
		},
		genEnvMap,
	))

	properties.TestingRun(t)
}

// **Feature: env-audit, Property 4: Sensitive value redaction**
// **Validates: Requirements 2.2, 2.3, 8.2**
// For any environment map containing sensitive keys, FormatEnv output SHALL NOT
// contain the actual values of sensitive keys, and SHALL contain "[REDACTED]" for
// each sensitive key.
func TestProperty_SensitiveValueRedaction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	sensitivePatterns := []string{"SECRET", "PASSWORD", "TOKEN", "API_KEY", "CREDENTIAL"}

	// Generator for sensitive keys
	genSensitiveKey := gen.AlphaString().Map(func(prefix string) string {
		pattern := sensitivePatterns[len(prefix)%len(sensitivePatterns)]
		return prefix + "_" + pattern
	})

	// Generator for non-empty secret values - use longer values to avoid substring false positives
	genSecretValue := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 8 && s != "[REDACTED]"
	})

	// Generator for maps with at least one sensitive key
	genEnvWithSensitive := gen.MapOf(genSensitiveKey, genSecretValue).SuchThat(func(m map[string]string) bool {
		return len(m) > 0
	})

	properties.Property("sensitive values are never in output and [REDACTED] appears", prop.ForAll(
		func(env map[string]string) bool {
			output := FormatEnv(env, true)

			for key, value := range env {
				if audit.IsSensitiveKey(key) {
					// Value should NOT appear in output (only check if value is long enough to be meaningful)
					if strings.Contains(output, value) {
						return false
					}
					// [REDACTED] should appear for this key
					if !strings.Contains(output, key+"=[REDACTED]") {
						return false
					}
				}
			}

			return true
		},
		genEnvWithSensitive,
	))

	properties.TestingRun(t)
}
