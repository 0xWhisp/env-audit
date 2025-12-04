package parser

import (
	"strings"
	"testing"

	"env-audit/internal/audit"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit-v2, Property 9: Diff categorization correctness**
// **Validates: Requirements 8.1, 8.2**
// For any two environment maps, Diff SHALL correctly categorize:
// - Added: keys in file2 but not file1
// - Removed: keys in file1 but not file2
// - Changed: keys in both with different values
func TestProperty_DiffCategorizationCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid key names
	genKey := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})

	// Generator for values
	genValue := gen.AlphaString()

	// Generator for env maps
	genEnvMap := gen.MapOf(genKey, genValue)

	properties.Property("Added contains only keys in file2 but not file1", prop.ForAll(
		func(file1, file2 map[string]string) bool {
			result := Diff(file1, file2)

			for key := range result.Added {
				// Key must be in file2
				if _, inFile2 := file2[key]; !inFile2 {
					return false
				}
				// Key must NOT be in file1
				if _, inFile1 := file1[key]; inFile1 {
					return false
				}
			}

			// All keys in file2 but not file1 must be in Added
			for key := range file2 {
				if _, inFile1 := file1[key]; !inFile1 {
					if _, inAdded := result.Added[key]; !inAdded {
						return false
					}
				}
			}

			return true
		},
		genEnvMap,
		genEnvMap,
	))

	properties.Property("Removed contains only keys in file1 but not file2", prop.ForAll(
		func(file1, file2 map[string]string) bool {
			result := Diff(file1, file2)

			for key := range result.Removed {
				// Key must be in file1
				if _, inFile1 := file1[key]; !inFile1 {
					return false
				}
				// Key must NOT be in file2
				if _, inFile2 := file2[key]; inFile2 {
					return false
				}
			}

			// All keys in file1 but not file2 must be in Removed
			for key := range file1 {
				if _, inFile2 := file2[key]; !inFile2 {
					if _, inRemoved := result.Removed[key]; !inRemoved {
						return false
					}
				}
			}

			return true
		},
		genEnvMap,
		genEnvMap,
	))

	properties.Property("Changed contains only keys in both with different values", prop.ForAll(
		func(file1, file2 map[string]string) bool {
			result := Diff(file1, file2)

			for key, values := range result.Changed {
				// Key must be in both maps
				val1, inFile1 := file1[key]
				val2, inFile2 := file2[key]
				if !inFile1 || !inFile2 {
					return false
				}
				// Values must be different
				if val1 == val2 {
					return false
				}
				// Changed values must match
				if values[0] != val1 || values[1] != val2 {
					return false
				}
			}

			// All keys in both with different values must be in Changed
			for key, val1 := range file1 {
				if val2, inFile2 := file2[key]; inFile2 && val1 != val2 {
					if _, inChanged := result.Changed[key]; !inChanged {
						return false
					}
				}
			}

			return true
		},
		genEnvMap,
		genEnvMap,
	))

	properties.TestingRun(t)
}

// **Feature: env-audit-v2, Property 10: Diff value redaction**
// **Validates: Requirements 8.3**
// For any diff containing sensitive keys, FormatDiff with redact=true SHALL
// replace sensitive values with [REDACTED] and never expose actual values.
func TestProperty_DiffValueRedaction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for sensitive key patterns
	genSensitiveKey := gen.OneGenOf(
		gen.Const("API_KEY"),
		gen.Const("SECRET_TOKEN"),
		gen.Const("DB_PASSWORD"),
		gen.Const("AUTH_SECRET"),
		gen.Const("PRIVATE_KEY"),
		gen.Const("AWS_SECRET_KEY"),
	)

	// Generator for non-sensitive key patterns
	genNonSensitiveKey := gen.OneGenOf(
		gen.Const("APP_NAME"),
		gen.Const("DEBUG"),
		gen.Const("PORT"),
		gen.Const("HOST"),
		gen.Const("DATABASE_URL"),
	)

	// Generator for secret values
	genSecretValue := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 5
	})

	properties.Property("Sensitive values are redacted in Added entries", prop.ForAll(
		func(key, value string) bool {
			file1 := map[string]string{}
			file2 := map[string]string{key: value}

			result := Diff(file1, file2)
			output := FormatDiff(result, true)

			// If key is sensitive, value should NOT appear in output
			if audit.IsSensitiveKey(key) {
				if strings.Contains(output, value) && value != "" {
					return false
				}
				if !strings.Contains(output, "[REDACTED]") {
					return false
				}
			}
			return true
		},
		genSensitiveKey,
		genSecretValue,
	))

	properties.Property("Sensitive values are redacted in Removed entries", prop.ForAll(
		func(key, value string) bool {
			file1 := map[string]string{key: value}
			file2 := map[string]string{}

			result := Diff(file1, file2)
			output := FormatDiff(result, true)

			// If key is sensitive, value should NOT appear in output
			if audit.IsSensitiveKey(key) {
				if strings.Contains(output, value) && value != "" {
					return false
				}
				if !strings.Contains(output, "[REDACTED]") {
					return false
				}
			}
			return true
		},
		genSensitiveKey,
		genSecretValue,
	))

	properties.Property("Sensitive values are redacted in Changed entries", prop.ForAll(
		func(key, oldVal, newVal string) bool {
			if oldVal == newVal {
				return true // Skip if values are the same (no change)
			}
			file1 := map[string]string{key: oldVal}
			file2 := map[string]string{key: newVal}

			result := Diff(file1, file2)
			output := FormatDiff(result, true)

			// If key is sensitive, neither value should appear in output
			if audit.IsSensitiveKey(key) {
				if strings.Contains(output, oldVal) && oldVal != "" {
					return false
				}
				if strings.Contains(output, newVal) && newVal != "" {
					return false
				}
				if !strings.Contains(output, "[REDACTED]") {
					return false
				}
			}
			return true
		},
		genSensitiveKey,
		genSecretValue,
		genSecretValue,
	))

	properties.Property("Non-sensitive values are NOT redacted when redact=true", prop.ForAll(
		func(key, value string) bool {
			if value == "" {
				return true
			}
			file1 := map[string]string{}
			file2 := map[string]string{key: value}

			result := Diff(file1, file2)
			output := FormatDiff(result, true)

			// Non-sensitive key values should appear in output
			if !audit.IsSensitiveKey(key) {
				if !strings.Contains(output, value) {
					return false
				}
			}
			return true
		},
		genNonSensitiveKey,
		genSecretValue,
	))

	properties.Property("No redaction when redact=false", prop.ForAll(
		func(key, value string) bool {
			if value == "" {
				return true
			}
			file1 := map[string]string{}
			file2 := map[string]string{key: value}

			result := Diff(file1, file2)
			output := FormatDiff(result, false)

			// Value should appear in output when redact is false
			if !strings.Contains(output, value) {
				return false
			}
			return true
		},
		genSensitiveKey,
		genSecretValue,
	))

	properties.TestingRun(t)
}

// Unit tests for Diff edge cases
func TestDiff_EmptyMaps(t *testing.T) {
	result := Diff(map[string]string{}, map[string]string{})
	if len(result.Added) != 0 {
		t.Errorf("expected 0 added, got %d", len(result.Added))
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(result.Removed))
	}
	if len(result.Changed) != 0 {
		t.Errorf("expected 0 changed, got %d", len(result.Changed))
	}
}

func TestDiff_IdenticalMaps(t *testing.T) {
	file1 := map[string]string{"A": "1", "B": "2"}
	file2 := map[string]string{"A": "1", "B": "2"}
	result := Diff(file1, file2)
	if len(result.Added) != 0 || len(result.Removed) != 0 || len(result.Changed) != 0 {
		t.Error("identical maps should have no diff")
	}
}

func TestDiff_AllAdded(t *testing.T) {
	file1 := map[string]string{}
	file2 := map[string]string{"A": "1", "B": "2"}
	result := Diff(file1, file2)
	if len(result.Added) != 2 {
		t.Errorf("expected 2 added, got %d", len(result.Added))
	}
	if len(result.Removed) != 0 || len(result.Changed) != 0 {
		t.Error("should only have added entries")
	}
}

func TestDiff_AllRemoved(t *testing.T) {
	file1 := map[string]string{"A": "1", "B": "2"}
	file2 := map[string]string{}
	result := Diff(file1, file2)
	if len(result.Removed) != 2 {
		t.Errorf("expected 2 removed, got %d", len(result.Removed))
	}
	if len(result.Added) != 0 || len(result.Changed) != 0 {
		t.Error("should only have removed entries")
	}
}

func TestDiff_AllChanged(t *testing.T) {
	file1 := map[string]string{"A": "1", "B": "2"}
	file2 := map[string]string{"A": "10", "B": "20"}
	result := Diff(file1, file2)
	if len(result.Changed) != 2 {
		t.Errorf("expected 2 changed, got %d", len(result.Changed))
	}
	if len(result.Added) != 0 || len(result.Removed) != 0 {
		t.Error("should only have changed entries")
	}
}

func TestDiff_Mixed(t *testing.T) {
	file1 := map[string]string{"A": "1", "B": "2"}
	file2 := map[string]string{"B": "20", "C": "3"}
	result := Diff(file1, file2)

	// A removed
	if _, ok := result.Removed["A"]; !ok {
		t.Error("A should be in removed")
	}
	// B changed
	if _, ok := result.Changed["B"]; !ok {
		t.Error("B should be in changed")
	}
	// C added
	if _, ok := result.Added["C"]; !ok {
		t.Error("C should be in added")
	}
}

func TestFormatDiff_Empty(t *testing.T) {
	result := &DiffResult{
		Added:   map[string]string{},
		Removed: map[string]string{},
		Changed: map[string][2]string{},
	}
	output := FormatDiff(result, true)
	if output != "" {
		t.Errorf("expected empty output for empty diff, got %q", output)
	}
}

func TestFormatDiff_Nil(t *testing.T) {
	output := FormatDiff(nil, true)
	if output != "" {
		t.Errorf("expected empty output for nil diff, got %q", output)
	}
}

func TestFormatDiff_Prefixes(t *testing.T) {
	result := &DiffResult{
		Added:   map[string]string{"NEW": "value"},
		Removed: map[string]string{"OLD": "value"},
		Changed: map[string][2]string{"MOD": {"old", "new"}},
	}
	output := FormatDiff(result, false)

	if !strings.Contains(output, "+ NEW=value") {
		t.Errorf("expected '+ NEW=value', got %q", output)
	}
	if !strings.Contains(output, "- OLD=value") {
		t.Errorf("expected '- OLD=value', got %q", output)
	}
	if !strings.Contains(output, "~ MOD=old -> new") {
		t.Errorf("expected '~ MOD=old -> new', got %q", output)
	}
}

func TestFormatDiff_Redaction(t *testing.T) {
	result := &DiffResult{
		Added:   map[string]string{"API_KEY": "secret123"},
		Removed: map[string]string{"DB_PASSWORD": "pass456"},
		Changed: map[string][2]string{"SECRET_TOKEN": {"old_secret", "new_secret"}},
	}
	output := FormatDiff(result, true)

	// Should not contain actual values
	if strings.Contains(output, "secret123") {
		t.Error("should not contain secret123")
	}
	if strings.Contains(output, "pass456") {
		t.Error("should not contain pass456")
	}
	if strings.Contains(output, "old_secret") || strings.Contains(output, "new_secret") {
		t.Error("should not contain secret values")
	}

	// Should contain [REDACTED]
	if !strings.Contains(output, "[REDACTED]") {
		t.Error("should contain [REDACTED]")
	}
}

func TestFormatDiff_NonSensitiveNotRedacted(t *testing.T) {
	result := &DiffResult{
		Added: map[string]string{"DEBUG": "true"},
	}
	output := FormatDiff(result, true)

	if !strings.Contains(output, "true") {
		t.Error("non-sensitive value should not be redacted")
	}
}

