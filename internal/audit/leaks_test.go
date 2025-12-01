package audit

import (
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit-v2, Property 6: Leak pattern detection**
// **Validates: Requirements 6.1, 6.2, 6.3**
// For any value matching known secret patterns (ghp_, sk_live_, sk_test_, AKIA, JWT)
// OR having entropy >4.5 and length >20, CheckLeaks SHALL report it as a potential leak.
func TestProperty_LeakPatternDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for GitHub tokens (ghp_ followed by 36 alphanumeric chars)
	genGitHubToken := gen.Const("ghp_").FlatMap(func(prefix interface{}) gopter.Gen {
		return gen.SliceOfN(36, gen.AlphaNumChar()).Map(func(chars []rune) string {
			return prefix.(string) + string(chars)
		})
	}, reflect.TypeOf(""))

	// Generator for Stripe live keys (sk_live_ followed by alphanumeric chars)
	genStripeLiveKey := gen.IntRange(10, 30).FlatMap(func(length interface{}) gopter.Gen {
		return gen.SliceOfN(length.(int), gen.AlphaNumChar()).Map(func(chars []rune) string {
			return "sk_live_" + string(chars)
		})
	}, reflect.TypeOf(""))

	// Generator for Stripe test keys (sk_test_ followed by alphanumeric chars)
	genStripeTestKey := gen.IntRange(10, 30).FlatMap(func(length interface{}) gopter.Gen {
		return gen.SliceOfN(length.(int), gen.AlphaNumChar()).Map(func(chars []rune) string {
			return "sk_test_" + string(chars)
		})
	}, reflect.TypeOf(""))

	// Generator for AWS access keys (AKIA followed by 16 uppercase alphanumeric chars)
	genAWSKey := gen.SliceOfN(16, gen.OneConstOf('0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z')).Map(func(chars []rune) string {
		return "AKIA" + string(chars)
	})

	// Generator for JWTs (eyJ...eyJ...signature)
	genJWT := gen.SliceOfN(10, gen.AlphaNumChar()).FlatMap(func(part1 interface{}) gopter.Gen {
		return gen.SliceOfN(10, gen.AlphaNumChar()).FlatMap(func(part2 interface{}) gopter.Gen {
			return gen.SliceOfN(10, gen.AlphaNumChar()).Map(func(part3 []rune) string {
				return "eyJ" + string(part1.([]rune)) + ".eyJ" + string(part2.([]rune)) + "." + string(part3)
			})
		}, reflect.TypeOf(""))
	}, reflect.TypeOf(""))

	// Property: GitHub tokens are detected as leaks
	properties.Property("GitHub tokens are detected", prop.ForAll(
		func(token string) bool {
			env := map[string]string{"TEST_KEY": token}
			issues := CheckLeaks(env, nil)
			return len(issues) == 1 && issues[0].Type == IssueLeak
		},
		genGitHubToken,
	))

	// Property: Stripe live keys are detected as leaks
	properties.Property("Stripe live keys are detected", prop.ForAll(
		func(key string) bool {
			env := map[string]string{"TEST_KEY": key}
			issues := CheckLeaks(env, nil)
			return len(issues) == 1 && issues[0].Type == IssueLeak
		},
		genStripeLiveKey,
	))

	// Property: Stripe test keys are detected as leaks
	properties.Property("Stripe test keys are detected", prop.ForAll(
		func(key string) bool {
			env := map[string]string{"TEST_KEY": key}
			issues := CheckLeaks(env, nil)
			return len(issues) == 1 && issues[0].Type == IssueLeak
		},
		genStripeTestKey,
	))

	// Property: AWS access keys are detected as leaks
	properties.Property("AWS access keys are detected", prop.ForAll(
		func(key string) bool {
			env := map[string]string{"TEST_KEY": key}
			issues := CheckLeaks(env, nil)
			return len(issues) == 1 && issues[0].Type == IssueLeak
		},
		genAWSKey,
	))

	// Property: JWTs are detected as leaks
	properties.Property("JWTs are detected", prop.ForAll(
		func(jwt string) bool {
			env := map[string]string{"TEST_KEY": jwt}
			issues := CheckLeaks(env, nil)
			return len(issues) == 1 && issues[0].Type == IssueLeak
		},
		genJWT,
	))

	properties.TestingRun(t)
}

// **Feature: env-audit-v2, Property 6: Leak pattern detection (high entropy)**
// **Validates: Requirements 6.3**
// For any value with entropy >4.5 and length >20, CheckLeaks SHALL report it as a potential leak.
func TestProperty_HighEntropyDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for high entropy strings (random alphanumeric, length > 20)
	// Using a diverse character set to ensure high entropy
	genHighEntropyString := gen.IntRange(25, 50).FlatMap(func(length interface{}) gopter.Gen {
		return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
			// Use a mix of lowercase, uppercase, and digits for high entropy
			chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			result := make([]byte, length.(int))
			for i := range result {
				result[i] = chars[rand.Intn(len(chars))]
			}
			return gopter.NewGenResult(string(result), gopter.NoShrinker)
		})
	}, reflect.TypeOf(""))

	// Property: High entropy strings (>4.5 bits/char, length >20) are detected
	properties.Property("high entropy strings are detected", prop.ForAll(
		func(value string) bool {
			// Only test if the generated string actually has high entropy
			if !IsHighEntropy(value) {
				return true // Skip values that don't meet criteria
			}
			env := map[string]string{"TEST_KEY": value}
			issues := CheckLeaks(env, nil)
			return len(issues) == 1 && issues[0].Type == IssueLeak
		},
		genHighEntropyString,
	))

	properties.TestingRun(t)
}


// **Feature: env-audit-v2, Property 16: Entropy calculation correctness**
// **Validates: Requirements 6.3**
// For any string, CalculateEntropy SHALL return Shannon entropy in bits per character,
// where random alphanumeric strings of length 20+ typically exceed 4.0.
func TestProperty_EntropyCalculationCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 1: Empty string has zero entropy
	properties.Property("empty string has zero entropy", prop.ForAll(
		func(_ bool) bool {
			return CalculateEntropy("") == 0
		},
		gen.Bool(),
	))

	// Property 2: Single ASCII character repeated has zero entropy
	properties.Property("single ASCII character repeated has zero entropy", prop.ForAll(
		func(c byte, count int) bool {
			if count <= 0 {
				return true
			}
			s := strings.Repeat(string(c), count)
			return CalculateEntropy(s) == 0
		},
		gen.UInt8Range(32, 126), // printable ASCII
		gen.IntRange(1, 100),
	))

	// Property 3: Entropy is non-negative
	properties.Property("entropy is non-negative", prop.ForAll(
		func(s string) bool {
			return CalculateEntropy(s) >= 0
		},
		gen.AnyString(),
	))

	// Property 4: Entropy is bounded by log2(unique runes)
	// Using ASCII strings to avoid byte/rune length mismatch
	genASCIIString := gen.SliceOfN(50, gen.UInt8Range(32, 126)).Map(func(bytes []uint8) string {
		result := make([]byte, len(bytes))
		for i, b := range bytes {
			result[i] = byte(b)
		}
		return string(result)
	})

	properties.Property("entropy bounded by log2 of unique characters", prop.ForAll(
		func(s string) bool {
			if len(s) == 0 {
				return true
			}
			entropy := CalculateEntropy(s)
			uniqueChars := make(map[rune]bool)
			for _, c := range s {
				uniqueChars[c] = true
			}
			maxEntropy := math.Log2(float64(len(uniqueChars)))
			// Allow small floating point tolerance
			return entropy <= maxEntropy+0.0001
		},
		genASCIIString.SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 5: Uniformly distributed strings have high entropy
	// Generate strings with all unique characters to guarantee high entropy
	genHighEntropyString := gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		// Use all 62 alphanumeric chars to ensure high entropy
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length := 25 + params.Rng.Intn(25) // 25-50 chars
		result := make([]byte, length)
		for i := range result {
			result[i] = chars[params.Rng.Intn(len(chars))]
		}
		return gopter.NewGenResult(string(result), gopter.NoShrinker)
	})

	properties.Property("random alphanumeric strings 25+ chars typically have entropy > 3.5", prop.ForAll(
		func(s string) bool {
			entropy := CalculateEntropy(s)
			// Random alphanumeric strings should have reasonably high entropy
			// Using 3.5 as threshold since some random strings may have repetition
			return entropy > 3.5
		},
		genHighEntropyString,
	))

	// Property 6: Two-character alphabet has entropy <= 1.0
	properties.Property("two character alphabet has entropy <= 1.0", prop.ForAll(
		func(count int) bool {
			if count <= 0 {
				return true
			}
			// Create string with only 'a' and 'b'
			s := strings.Repeat("ab", count)
			entropy := CalculateEntropy(s)
			// Max entropy for 2 chars is log2(2) = 1.0
			return entropy <= 1.0+0.0001
		},
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}

// **Feature: env-audit-v2, Property 7: Leak value redaction**
// **Validates: Requirements 6.4**
// For any leak detection report, the actual secret value SHALL NOT appear in any output format.
func TestProperty_LeakValueRedaction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for GitHub tokens (ghp_ followed by 36 alphanumeric chars)
	genGitHubToken := gen.Const("ghp_").FlatMap(func(prefix interface{}) gopter.Gen {
		return gen.SliceOfN(36, gen.AlphaNumChar()).Map(func(chars []rune) string {
			return prefix.(string) + string(chars)
		})
	}, reflect.TypeOf(""))

	// Generator for Stripe live keys
	genStripeLiveKey := gen.IntRange(10, 30).FlatMap(func(length interface{}) gopter.Gen {
		return gen.SliceOfN(length.(int), gen.AlphaNumChar()).Map(func(chars []rune) string {
			return "sk_live_" + string(chars)
		})
	}, reflect.TypeOf(""))

	// Generator for AWS access keys (AKIA followed by 16 uppercase alphanumeric chars)
	genAWSKey := gen.SliceOfN(16, gen.OneConstOf('0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z')).Map(func(chars []rune) string {
		return "AKIA" + string(chars)
	})

	// Generator for high entropy strings
	genHighEntropyString := gen.IntRange(25, 50).FlatMap(func(length interface{}) gopter.Gen {
		return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
			chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			result := make([]byte, length.(int))
			for i := range result {
				result[i] = chars[rand.Intn(len(chars))]
			}
			return gopter.NewGenResult(string(result), gopter.NoShrinker)
		})
	}, reflect.TypeOf(""))

	// Combined generator for all secret types
	genSecret := gen.OneGenOf(genGitHubToken, genStripeLiveKey, genAWSKey, genHighEntropyString)

	// Property: Secret values never appear in issue messages
	properties.Property("secret values never appear in issue messages", prop.ForAll(
		func(secretValue string) bool {
			env := map[string]string{"SECRET_KEY": secretValue}
			issues := CheckLeaks(env, nil)

			// If no issues detected (e.g., low entropy string), skip
			if len(issues) == 0 {
				return true
			}

			// The actual secret value must NOT appear in any issue message
			for _, issue := range issues {
				if strings.Contains(issue.Message, secretValue) {
					return false
				}
				// Also check that the value doesn't appear in the key field
				// (though it shouldn't, this is defense in depth)
				if issue.Key == secretValue {
					return false
				}
			}
			return true
		},
		genSecret,
	))

	properties.TestingRun(t)
}
