package main

import (
	"fmt"
	"strings"
)

// FormatSummary produces human-readable output grouped by issue type
func FormatSummary(result *ScanResult) string {
	if result == nil || len(result.Issues) == 0 {
		return "env-audit scan results\n======================\n\nNo issues found.\n"
	}

	// Group issues by type
	groups := make(map[IssueType][]Issue)
	for _, issue := range result.Issues {
		groups[issue.Type] = append(groups[issue.Type], issue)
	}

	var sb strings.Builder
	sb.WriteString("env-audit scan results\n")
	sb.WriteString("======================\n")

	// Output each group in order
	typeOrder := []IssueType{IssueEmpty, IssueMissing, IssueSensitive, IssueDuplicate}
	typeNames := map[IssueType]string{
		IssueEmpty:     "Empty Values",
		IssueMissing:   "Missing Required",
		IssueSensitive: "Sensitive Keys Detected",
		IssueDuplicate: "Duplicate Keys",
	}

	for _, t := range typeOrder {
		issues := groups[t]
		if len(issues) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s (%d):\n", typeNames[t], len(issues)))
		for _, issue := range issues {
			if t == IssueSensitive {
				sb.WriteString(fmt.Sprintf("  - %s: [REDACTED]\n", issue.Key))
			} else {
				sb.WriteString(fmt.Sprintf("  - %s\n", issue.Key))
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d issues found\n", len(result.Issues)))
	return sb.String()
}

// Redact returns "[REDACTED]" placeholder
func Redact(value string) string {
	// TODO: Implement in task 6
	return "[REDACTED]"
}
