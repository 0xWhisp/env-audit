package parser

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit-v2, Property 5: Example comparison completeness**
// **Validates: Requirements 5.1, 5.2, 5.3**
// For any two environment maps (target and example), Compare SHALL return all keys
// in example but not in target as Missing, and all keys in target but not in example as Extra.
func TestProperty_ExampleComparisonCompleteness(t *testing.T) {
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

	properties.Property("Compare returns all missing and extra keys correctly", prop.ForAll(
		func(target, example map[string]string) bool {
			result := Compare(target, example)

			// Calculate expected missing (in example but not in target)
			expectedMissing := make(map[string]bool)
			for key := range example {
				if _, exists := target[key]; !exists {
					expectedMissing[key] = true
				}
			}

			// Calculate expected extra (in target but not in example)
			expectedExtra := make(map[string]bool)
			for key := range target {
				if _, exists := example[key]; !exists {
					expectedExtra[key] = true
				}
			}

			// Verify missing keys
			if len(result.Missing) != len(expectedMissing) {
				return false
			}
			for _, key := range result.Missing {
				if !expectedMissing[key] {
					return false
				}
			}

			// Verify extra keys
			if len(result.Extra) != len(expectedExtra) {
				return false
			}
			for _, key := range result.Extra {
				if !expectedExtra[key] {
					return false
				}
			}

			return true
		},
		genEnvMap,
		genEnvMap,
	))

	properties.TestingRun(t)
}

// Unit tests for Compare edge cases
func TestCompare_EmptyMaps(t *testing.T) {
	result := Compare(map[string]string{}, map[string]string{})
	if len(result.Missing) != 0 {
		t.Errorf("expected 0 missing, got %d", len(result.Missing))
	}
	if len(result.Extra) != 0 {
		t.Errorf("expected 0 extra, got %d", len(result.Extra))
	}
}

func TestCompare_IdenticalMaps(t *testing.T) {
	target := map[string]string{"A": "1", "B": "2"}
	example := map[string]string{"A": "1", "B": "2"}
	result := Compare(target, example)
	if len(result.Missing) != 0 {
		t.Errorf("expected 0 missing, got %d", len(result.Missing))
	}
	if len(result.Extra) != 0 {
		t.Errorf("expected 0 extra, got %d", len(result.Extra))
	}
}

func TestCompare_AllMissing(t *testing.T) {
	target := map[string]string{}
	example := map[string]string{"A": "1", "B": "2"}
	result := Compare(target, example)
	if len(result.Missing) != 2 {
		t.Errorf("expected 2 missing, got %d", len(result.Missing))
	}
	if len(result.Extra) != 0 {
		t.Errorf("expected 0 extra, got %d", len(result.Extra))
	}
}

func TestCompare_AllExtra(t *testing.T) {
	target := map[string]string{"A": "1", "B": "2"}
	example := map[string]string{}
	result := Compare(target, example)
	if len(result.Missing) != 0 {
		t.Errorf("expected 0 missing, got %d", len(result.Missing))
	}
	if len(result.Extra) != 2 {
		t.Errorf("expected 2 extra, got %d", len(result.Extra))
	}
}

func TestCompare_MixedDifferences(t *testing.T) {
	target := map[string]string{"A": "1", "C": "3"}
	example := map[string]string{"A": "1", "B": "2"}
	result := Compare(target, example)

	// B is in example but not target -> missing
	if len(result.Missing) != 1 || result.Missing[0] != "B" {
		t.Errorf("expected missing=[B], got %v", result.Missing)
	}

	// C is in target but not example -> extra
	if len(result.Extra) != 1 || result.Extra[0] != "C" {
		t.Errorf("expected extra=[C], got %v", result.Extra)
	}
}
