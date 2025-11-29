package main

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
			issues := CheckEmpty(env)

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
			issues := CheckMissing(env, required)

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
