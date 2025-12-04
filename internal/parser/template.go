package parser

import (
	"sort"
	"strings"

	"env-audit/internal/audit"
)

// GenerateTemplate creates .env.example content from an environment map.
// Sensitive keys get empty values, non-sensitive keys get placeholder values.
func GenerateTemplate(env map[string]string) string {
	if len(env) == 0 {
		return ""
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var lines []string
	for _, key := range keys {
		if audit.IsSensitiveKey(key) {
			lines = append(lines, key+"=")
		} else {
			lines = append(lines, key+"=your_"+strings.ToLower(key)+"_here")
		}
	}

	return strings.Join(lines, "\n")
}
