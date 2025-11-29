package main

import "strings"

// IssueType represents the category of an audit issue
type IssueType int

const (
	IssueEmpty IssueType = iota
	IssueMissing
	IssueSensitive
	IssueDuplicate
)

// Issue represents a single audit finding
type Issue struct {
	Type    IssueType
	Key     string
	Message string
}

// CheckEmpty finds variables with empty values
func CheckEmpty(env map[string]string) []Issue {
	var issues []Issue
	for key, value := range env {
		if value == "" {
			issues = append(issues, Issue{
				Type:    IssueEmpty,
				Key:     key,
				Message: "variable has empty value",
			})
		}
	}
	return issues
}

// CheckMissing finds required variables not present
func CheckMissing(env map[string]string, required []string) []Issue {
	var issues []Issue
	seen := make(map[string]bool)
	for _, key := range required {
		if seen[key] {
			continue
		}
		seen[key] = true
		if _, exists := env[key]; !exists {
			issues = append(issues, Issue{
				Type:    IssueMissing,
				Key:     key,
				Message: "required variable is missing",
			})
		}
	}
	return issues
}

// CheckSensitive finds keys matching sensitive patterns
func CheckSensitive(env map[string]string) []Issue {
	var issues []Issue
	for key := range env {
		if IsSensitiveKey(key) {
			issues = append(issues, Issue{
				Type:    IssueSensitive,
				Key:     key,
				Message: "sensitive key detected",
			})
		}
	}
	return issues
}

// IsSensitiveKey returns true if key matches sensitive patterns
// Matches: SECRET, PASSWORD, TOKEN, API_KEY, APIKEY, KEY suffix, CREDENTIAL, PRIVATE, AUTH
func IsSensitiveKey(key string) bool {
	upper := strings.ToUpper(key)

	// Check for exact patterns contained anywhere in the key
	patterns := []string{"SECRET", "PASSWORD", "TOKEN", "API_KEY", "APIKEY", "CREDENTIAL", "PRIVATE", "AUTH"}
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}

	// Check for KEY suffix (e.g., STRIPE_KEY, AWS_KEY)
	if strings.HasSuffix(upper, "KEY") {
		return true
	}

	return false
}
