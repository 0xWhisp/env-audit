package parser

import (
	"sort"
	"strings"

	"env-audit/internal/audit"
)

// DiffResult contains the differences between two env files
type DiffResult struct {
	Added   map[string]string    // keys in file2 but not in file1
	Removed map[string]string    // keys in file1 but not in file2
	Changed map[string][2]string // keys with different values [old, new]
}

// Diff compares two environment maps and returns their differences
func Diff(file1, file2 map[string]string) *DiffResult {
	result := &DiffResult{
		Added:   make(map[string]string),
		Removed: make(map[string]string),
		Changed: make(map[string][2]string),
	}

	// Find removed keys (in file1 but not in file2)
	for key, val1 := range file1 {
		if val2, exists := file2[key]; !exists {
			result.Removed[key] = val1
		} else if val1 != val2 {
			result.Changed[key] = [2]string{val1, val2}
		}
	}

	// Find added keys (in file2 but not in file1)
	for key, val2 := range file2 {
		if _, exists := file1[key]; !exists {
			result.Added[key] = val2
		}
	}

	return result
}

// FormatDiff formats a DiffResult as a human-readable string with +/- prefixes.
// If redact is true, sensitive values are replaced with [REDACTED].
func FormatDiff(result *DiffResult, redact bool) string {
	if result == nil {
		return ""
	}

	var lines []string

	// Collect and sort keys for consistent output
	addedKeys := sortedKeys(result.Added)
	removedKeys := sortedKeys(result.Removed)
	changedKeys := sortedKeysFromChanged(result.Changed)

	// Format removed lines (-)
	for _, key := range removedKeys {
		val := redactValue(key, result.Removed[key], redact)
		lines = append(lines, "- "+key+"="+val)
	}

	// Format added lines (+)
	for _, key := range addedKeys {
		val := redactValue(key, result.Added[key], redact)
		lines = append(lines, "+ "+key+"="+val)
	}

	// Format changed lines (~)
	for _, key := range changedKeys {
		oldVal := redactValue(key, result.Changed[key][0], redact)
		newVal := redactValue(key, result.Changed[key][1], redact)
		lines = append(lines, "~ "+key+"="+oldVal+" -> "+newVal)
	}

	return strings.Join(lines, "\n")
}

// redactValue returns [REDACTED] if redact is true and key is sensitive
func redactValue(key, value string, redact bool) string {
	if redact && audit.IsSensitiveKey(key) {
		return "[REDACTED]"
	}
	return value
}

// sortedKeys returns sorted keys from a map
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedKeysFromChanged returns sorted keys from a changed map
func sortedKeysFromChanged(m map[string][2]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
