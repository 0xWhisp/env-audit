package audit

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit, Property 3: Sensitive key pattern matching**
// **Validates: Requirements 2.1**
// For any key string containing one of the sensitive patterns (SECRET, PASSWORD, TOKEN,
// API_KEY, APIKEY, KEY suffix, CREDENTIAL, PRIVATE, AUTH) case-insensitively,
// IsSensitiveKey SHALL return true.
func TestProperty_SensitiveKeyPatternMatching(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	sensitivePatterns := []string{"SECRET", "PASSWORD", "TOKEN", "API_KEY", "APIKEY", "CREDENTIAL", "PRIVATE", "AUTH"}

	// Generator for keys that contain a sensitive pattern
	genSensitiveKey := gen.AnyString().Map(func(prefix string) string {
		// Pick a random pattern and embed it
		pattern := sensitivePatterns[len(prefix)%len(sensitivePatterns)]
		// Vary case randomly based on prefix length
		if len(prefix)%2 == 0 {
			pattern = strings.ToLower(pattern)
		}
		return prefix + pattern + "SUFFIX"
	})

	// Property: Keys containing sensitive patterns should be detected
	properties.Property("keys with sensitive patterns are detected", prop.ForAll(
		func(key string) bool {
			return IsSensitiveKey(key)
		},
		genSensitiveKey,
	))

	// Generator for keys ending with KEY suffix
	genKeySuffix := gen.AnyString().Map(func(prefix string) string {
		// Ensure prefix doesn't already contain sensitive patterns
		clean := strings.ReplaceAll(prefix, "SECRET", "")
		clean = strings.ReplaceAll(clean, "PASSWORD", "")
		clean = strings.ReplaceAll(clean, "TOKEN", "")
		clean = strings.ReplaceAll(clean, "CREDENTIAL", "")
		clean = strings.ReplaceAll(clean, "PRIVATE", "")
		clean = strings.ReplaceAll(clean, "AUTH", "")
		if len(clean)%2 == 0 {
			return clean + "KEY"
		}
		return clean + "key"
	})


	// Property: Keys ending with KEY suffix should be detected
	properties.Property("keys with KEY suffix are detected", prop.ForAll(
		func(key string) bool {
			return IsSensitiveKey(key)
		},
		genKeySuffix,
	))

	properties.TestingRun(t)
}

// **Feature: env-audit, Property 1: Empty value detection completeness**
// **Validates: Requirements 1.1**
// For any environment map, CheckEmpty SHALL return an issue for every key with an
// empty string value, and no issues for keys with non-empty values.
func TestProperty_EmptyValueDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for environment maps with known empty and non-empty values
	genEnvMap := gen.MapOf(gen.AlphaString(), gen.AnyString())

	properties.Property("detects all empty values and only empty values", prop.ForAll(
		func(env map[string]string) bool {
			issues := CheckEmpty(env, nil)

			// Count expected empty keys
			expectedEmpty := make(map[string]bool)
			for key, value := range env {
				if value == "" {
					expectedEmpty[key] = true
				}
			}

			// Check that we got exactly the right number of issues
			if len(issues) != len(expectedEmpty) {
				return false
			}

			// Check that all issues are for empty keys
			for _, issue := range issues {
				if issue.Type != IssueEmpty {
					return false
				}
				if !expectedEmpty[issue.Key] {
					return false
				}
			}

			return true
		},
		genEnvMap,
	))

	properties.TestingRun(t)
}


// **Feature: env-audit, Property 2: Missing required detection completeness**
// **Validates: Requirements 1.2**
// For any environment map and list of required variable names, CheckMissing SHALL
// return an issue for exactly those required names not present as keys in the map.
func TestProperty_MissingRequiredDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("detects exactly the missing required keys", prop.ForAll(
		func(env map[string]string, required []string) bool {
			issues := CheckMissing(env, required, nil)

			// Calculate expected missing keys
			expectedMissing := make(map[string]bool)
			for _, key := range required {
				if _, exists := env[key]; !exists {
					expectedMissing[key] = true
				}
			}

			// Check that we got exactly the right number of issues
			if len(issues) != len(expectedMissing) {
				return false
			}

			// Check that all issues are for missing keys
			for _, issue := range issues {
				if issue.Type != IssueMissing {
					return false
				}
				if !expectedMissing[issue.Key] {
					return false
				}
			}

			return true
		},
		gen.MapOf(gen.AlphaString(), gen.AnyString()),
		gen.SliceOf(gen.AlphaString()),
	))

	properties.TestingRun(t)
}

// Unit tests for edge cases
// Requirements: 1.1, 1.2, 2.1

func TestCheckEmpty_EmptyInput(t *testing.T) {
	issues := CheckEmpty(map[string]string{}, nil)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for empty input, got %d", len(issues))
	}
}

func TestCheckEmpty_SingleEntry(t *testing.T) {
	// Single empty entry
	issues := CheckEmpty(map[string]string{"FOO": ""}, nil)
	if len(issues) != 1 || issues[0].Key != "FOO" {
		t.Errorf("expected 1 issue for FOO, got %v", issues)
	}

	// Single non-empty entry
	issues = CheckEmpty(map[string]string{"FOO": "bar"}, nil)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for non-empty value, got %d", len(issues))
	}
}


func TestCheckEmpty_SpecialCharacters(t *testing.T) {
	env := map[string]string{
		"MY_VAR-1":    "",
		"VAR.NAME":    "",
		"VAR@SPECIAL": "value",
	}
	issues := CheckEmpty(env, nil)
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestCheckEmpty_WithIgnore(t *testing.T) {
	env := map[string]string{
		"FOO": "",
		"BAR": "",
	}
	issues := CheckEmpty(env, []string{"FOO"})
	if len(issues) != 1 || issues[0].Key != "BAR" {
		t.Errorf("expected 1 issue for BAR, got %v", issues)
	}
}

func TestCheckMissing_EmptyInput(t *testing.T) {
	// Empty env, empty required
	issues := CheckMissing(map[string]string{}, []string{}, nil)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}

	// Empty env, one required
	issues = CheckMissing(map[string]string{}, []string{"FOO"}, nil)
	if len(issues) != 1 || issues[0].Key != "FOO" {
		t.Errorf("expected 1 issue for FOO, got %v", issues)
	}
}

func TestCheckMissing_SingleEntry(t *testing.T) {
	env := map[string]string{"FOO": "bar"}

	// Required key exists
	issues := CheckMissing(env, []string{"FOO"}, nil)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues when key exists, got %d", len(issues))
	}

	// Required key missing
	issues = CheckMissing(env, []string{"BAR"}, nil)
	if len(issues) != 1 || issues[0].Key != "BAR" {
		t.Errorf("expected 1 issue for BAR, got %v", issues)
	}
}

func TestCheckMissing_SpecialCharacters(t *testing.T) {
	env := map[string]string{"MY_VAR-1": "val"}
	issues := CheckMissing(env, []string{"MY_VAR-1", "VAR.NAME"}, nil)
	if len(issues) != 1 || issues[0].Key != "VAR.NAME" {
		t.Errorf("expected 1 issue for VAR.NAME, got %v", issues)
	}
}

func TestCheckMissing_WithIgnore(t *testing.T) {
	env := map[string]string{}
	issues := CheckMissing(env, []string{"FOO", "BAR"}, []string{"FOO"})
	if len(issues) != 1 || issues[0].Key != "BAR" {
		t.Errorf("expected 1 issue for BAR, got %v", issues)
	}
}

func TestCheckSensitive_EmptyInput(t *testing.T) {
	issues := CheckSensitive(map[string]string{}, nil)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for empty input, got %d", len(issues))
	}
}

func TestCheckSensitive_SingleEntry(t *testing.T) {
	// Sensitive key
	issues := CheckSensitive(map[string]string{"API_KEY": "secret"}, nil)
	if len(issues) != 1 || issues[0].Key != "API_KEY" {
		t.Errorf("expected 1 issue for API_KEY, got %v", issues)
	}

	// Non-sensitive key
	issues = CheckSensitive(map[string]string{"APP_NAME": "myapp"}, nil)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for non-sensitive key, got %d", len(issues))
	}
}

func TestCheckSensitive_SpecialCharacters(t *testing.T) {
	env := map[string]string{
		"MY-SECRET-VAR": "val",
		"VAR.PASSWORD":  "val",
		"NORMAL@VAR":    "val",
	}
	issues := CheckSensitive(env, nil)
	if len(issues) != 2 {
		t.Errorf("expected 2 sensitive issues, got %d", len(issues))
	}
}

func TestCheckSensitive_WithIgnore(t *testing.T) {
	env := map[string]string{
		"API_KEY":    "secret",
		"API_SECRET": "hidden",
	}
	issues := CheckSensitive(env, []string{"API_KEY"})
	if len(issues) != 1 || issues[0].Key != "API_SECRET" {
		t.Errorf("expected 1 issue for API_SECRET, got %v", issues)
	}
}

func TestIsSensitiveKey_EdgeCases(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"", false},
		{"KEY", true},            // KEY suffix
		{"MYKEY", true},          // KEY suffix
		{"KEYRING", false},       // KEY not as suffix
		{"secret", true},         // lowercase
		{"PaSsWoRd", true},       // mixed case
		{"MY_API_KEY_VAR", true}, // API_KEY in middle
		{"AUTHENTICATE", true},   // contains AUTH
		{"AUTHOR", true},         // contains AUTH
		{"PRIVATE_DATA", true},   // contains PRIVATE
		{"CREDENTIAL_ID", true},  // contains CREDENTIAL
	}

	for _, tc := range tests {
		got := IsSensitiveKey(tc.key)
		if got != tc.expected {
			t.Errorf("IsSensitiveKey(%q) = %v, want %v", tc.key, got, tc.expected)
		}
	}
}
