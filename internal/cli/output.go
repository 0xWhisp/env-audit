package cli

import (
	"fmt"
	"io"
	"strings"

	"env-audit/internal/audit"
)

// FormatSummary produces human-readable output grouped by issue type
func FormatSummary(result *audit.Result) string {
	if result == nil || len(result.Issues) == 0 {
		return "env-audit scan results\n======================\n\nNo issues found.\n"
	}

	// Group issues by type
	groups := make(map[audit.IssueType][]audit.Issue)
	for _, issue := range result.Issues {
		groups[issue.Type] = append(groups[issue.Type], issue)
	}

	var sb strings.Builder
	sb.WriteString("env-audit scan results\n")
	sb.WriteString("======================\n")

	// Output each group in order
	typeOrder := []audit.IssueType{audit.IssueEmpty, audit.IssueMissing, audit.IssueSensitive, audit.IssueDuplicate}
	typeNames := map[audit.IssueType]string{
		audit.IssueEmpty:     "Empty Values",
		audit.IssueMissing:   "Missing Required",
		audit.IssueSensitive: "Sensitive Keys Detected",
		audit.IssueDuplicate: "Duplicate Keys",
	}

	for _, t := range typeOrder {
		issues := groups[t]
		if len(issues) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s (%d):\n", typeNames[t], len(issues)))
		for _, issue := range issues {
			if t == audit.IssueSensitive {
				sb.WriteString(fmt.Sprintf("  - %s: [REDACTED]\n", issue.Key))
			} else {
				sb.WriteString(fmt.Sprintf("  - %s\n", issue.Key))
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d issues found\n", len(result.Issues)))
	return sb.String()
}


// PrintUsage outputs help text
func PrintUsage(w io.Writer) {
	fmt.Fprintln(w, "env-audit [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --file, -f <path>     Path to .env file to scan (optional)")
	fmt.Fprintln(w, "  --required, -r <vars> Comma-separated list of required variables")
	fmt.Fprintln(w, "  --dump, -d            Output parsed configuration (with redaction)")
	fmt.Fprintln(w, "  --help, -h            Show this help message")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Exit Codes:")
	fmt.Fprintln(w, "  0  No risks found")
	fmt.Fprintln(w, "  1  Risks detected")
	fmt.Fprintln(w, "  2  Fatal error (invalid arguments, file not found)")
}

// Redact returns "[REDACTED]" placeholder
func Redact(value string) string {
	return "[REDACTED]"
}
