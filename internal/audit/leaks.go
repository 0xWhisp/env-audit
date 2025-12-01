package audit

import (
	"math"
	"regexp"
)

// LeakPattern defines a known secret pattern
type LeakPattern struct {
	Name    string
	Pattern *regexp.Regexp
}

// KnownPatterns contains patterns for detecting hardcoded secrets
var KnownPatterns = []LeakPattern{
	{"GitHub Token", regexp.MustCompile(`^ghp_[a-zA-Z0-9]{36}$`)},
	{"Stripe Live Key", regexp.MustCompile(`^sk_live_[a-zA-Z0-9]+$`)},
	{"Stripe Test Key", regexp.MustCompile(`^sk_test_[a-zA-Z0-9]+$`)},
	{"AWS Access Key", regexp.MustCompile(`^AKIA[0-9A-Z]{16}$`)},
	{"JWT", regexp.MustCompile(`^eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+$`)},
}

// MatchesLeakPattern checks if a value matches any known secret pattern
func MatchesLeakPattern(value string) (bool, string) {
	for _, lp := range KnownPatterns {
		if lp.Pattern.MatchString(value) {
			return true, lp.Name
		}
	}
	return false, ""
}

// CalculateEntropy computes Shannon entropy in bits per character
func CalculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	// Count character frequencies
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}

	// Calculate entropy
	length := float64(len(s))
	var entropy float64
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// IsHighEntropy returns true if the string has high entropy (>4.5 bits/char) and length >20
func IsHighEntropy(value string) bool {
	if len(value) <= 20 {
		return false
	}
	return CalculateEntropy(value) > 4.5
}
